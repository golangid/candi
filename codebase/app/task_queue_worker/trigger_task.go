package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

func (t *taskQueueWorker) triggerTask(workerIndex int) {

	runningTask, ok := workerIndexTask[workerIndex]
	if !ok {
		return
	}
	runningTask.activeInterval.Stop()

	semaphore[workerIndex-1] <- struct{}{}
	if t.isShutdown {
		return
	}

	t.wg.Add(1)
	go func(workerIndex int, task *Task) {
		defer func() {
			if r := recover(); r != nil {
				logger.LogRed("task_queue_worker > panic: " + t.ctx.Err().Error())
			}
			t.wg.Done()
			<-semaphore[workerIndex-1]
			refreshWorkerNotif <- struct{}{}
		}()

		if t.ctx.Err() != nil {
			logger.LogRed("task_queue_worker > ctx root err: " + t.ctx.Err().Error())
			return
		}

		t.execJob(task)

	}(workerIndex, runningTask)
}

func (t *taskQueueWorker) execJob(runningTask *Task) {
	jobID := queue.PopJob(t.ctx, runningTask.taskName)
	if jobID == "" {
		tryRegisterNextJob(t.ctx, runningTask.taskName)
		return
	}

	// lock for multiple worker (if running on multiple pods/instance)
	if defaultOption.locker.IsLocked(t.getLockKey(jobID)) {
		logger.LogYellow("task_queue_worker > job " + jobID + " is locked")
		return
	}
	defer defaultOption.locker.Unlock(t.getLockKey(jobID))

	var ctx context.Context
	ctx, runningTask.cancel = context.WithCancel(t.ctx)
	defer runningTask.cancel()

	selectedTask := registeredTask[runningTask.taskName]

	job, err := persistent.FindJobByID(ctx, jobID, "retry_histories")
	if err != nil || job.Status == string(statusStopped) {
		tryRegisterNextJob(ctx, runningTask.taskName)
		return
	}

	selectedHandler := selectedTask.handler
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	isRetry, startAt := false, time.Now()

	job.Retries++
	job.Status = string(statusRetrying)
	persistent.SaveJob(ctx, job)
	broadcastAllToSubscribers(t.ctx)

	if defaultOption.debugMode {
		log.Printf("\x1b[35;3mTask Queue Worker: executing task '%s' (job id: %s)\x1b[0m", job.TaskName, job.ID)
	}

	tryRegisterNextJob(ctx, runningTask.taskName)

	trace, ctx := tracer.StartTraceFromHeader(ctx, "TaskQueueWorker", map[string]string{})
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			trace.SetError(err)
			job.Error = err.Error()
			job.Status = string(statusFailure)
		}

		job.FinishedAt = time.Now()
		retryHistory := RetryHistory{
			Status: job.Status, Error: job.Error, TraceID: job.TraceID,
			StartAt: startAt, EndAt: job.FinishedAt,
			ErrorStack: job.ErrorStack,
		}

		trace.SetTag("is_retry", isRetry)
		if isRetry {
			job.Status = string(statusQueueing)
		}

		logger.LogGreen("task_queue > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish()

		persistent.SaveJob(t.ctx, job, retryHistory)
		broadcastAllToSubscribers(t.ctx)
	}()

	tags := trace.Tags()
	tags["job_id"], tags["task_name"], tags["retries"], tags["max_retry"] = job.ID, job.TaskName, job.Retries, job.MaxRetry
	trace.Log("job_args", job.Arguments)

	job.TraceID = tracer.GetTraceID(ctx)

	var eventContext candishared.EventContext
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(types.TaskQueue))
	eventContext.SetHandlerRoute(job.TaskName)
	eventContext.SetHeader(map[string]string{
		"retries":   strconv.Itoa(job.Retries),
		"max_retry": strconv.Itoa(job.MaxRetry),
		"interval":  job.Interval,
	})
	eventContext.SetKey(job.ID)
	eventContext.WriteString(job.Arguments)

	if len(selectedHandler.HandlerFuncs) == 0 {
		job.Error = "No handler found for exec this job"
		job.Status = string(statusFailure)
		return
	}

	err = selectedHandler.HandlerFuncs[0](&eventContext)

	if ctx.Err() != nil {
		job.Error = "Job has been stopped when running (context error: " + ctx.Err().Error() + ")."
		if err != nil {
			job.Error += " Error: " + err.Error()
		}
		job.Status = string(statusStopped)
		return
	}

	if err != nil {
		eventContext.SetError(err)
		trace.SetError(err)

		job.Error = err.Error()
		job.Status = string(statusFailure)

		switch e := err.(type) {
		case *candishared.ErrorRetrier:
			job.ErrorStack = e.StackTrace

			if job.Retries < job.MaxRetry && e.Delay > 0 {

				isRetry = true
				job.Interval = e.Delay.String()
				if e.NewRetryIntervalFunc != nil {
					job.Interval = e.NewRetryIntervalFunc(job.Retries).String()
				}

				// update job arguments if in error retry contains new args payload
				if len(e.NewArgsPayload) > 0 {
					job.Arguments = string(e.NewArgsPayload)
				}

				registerJobToWorker(job, selectedTask.workerIndex)
				queue.PushJob(ctx, job)
				return
			}

			logger.LogRed("TaskQueueWorker: GIVE UP: " + job.TaskName)
		}

	} else {
		job.Status = string(statusSuccess)
		job.Error = ""
	}

	for _, h := range selectedHandler.HandlerFuncs[1:] {
		h(&eventContext)
	}
}

func (t *taskQueueWorker) getLockKey(jobID string) string {
	return fmt.Sprintf("%s:task-queue-worker-lock:%s", t.service.Name(), jobID)
}

func tryRegisterNextJob(ctx context.Context, taskName string) {

	nextJobID := queue.NextJob(ctx, taskName)
	if nextJobID != "" {
		if nextJob, err := persistent.FindJobByID(ctx, nextJobID); err == nil {
			registerJobToWorker(nextJob, registeredTask[taskName].workerIndex)
		}
	} else {

		filter := Filter{
			Page: 1, Limit: 10,
			TaskName: taskName,
			Status:   []string{string(statusQueueing)},
		}
		count := persistent.CountAllJob(ctx, filter)

		if count == 0 {
			return
		}

		totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
		for filter.Page <= totalPages {
			for _, job := range persistent.FindAllJob(ctx, filter) {
				queue.PushJob(ctx, &job)
				registerJobToWorker(&job, registeredTask[job.TaskName].workerIndex)
			}
			filter.Page++
		}

	}

}
