package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

func (t *taskQueueWorker) triggerTask(workerIndex int) {

	runningTask, ok := runningWorkerIndexTask[workerIndex]
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
		var ctx context.Context
		ctx, runningTask.cancel = context.WithCancel(t.ctx)

		defer func() {
			if r := recover(); r != nil {
				logger.LogRed(fmt.Sprintf("task_queue_worker > panic: %v", r))
			}
			t.wg.Done()
			<-semaphore[workerIndex-1]
			runningTask.cancel()
			refreshWorkerNotif <- struct{}{}
		}()

		if t.ctx.Err() != nil {
			logger.LogRed("task_queue_worker > ctx root err: " + t.ctx.Err().Error())
			return
		}

		t.execJob(ctx, task)

	}(workerIndex, runningTask)
}

func (t *taskQueueWorker) execJob(ctx context.Context, runningTask *Task) {
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

	selectedTask := registeredTask[runningTask.taskName]

	job, err := persistent.FindJobByID(ctx, jobID, "retry_histories")
	if err != nil || job.Status != string(statusQueueing) {
		tryRegisterNextJob(ctx, runningTask.taskName)
		return
	}

	selectedHandler := selectedTask.handler
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	isRetry, startAt := false, time.Now()

	job.Retries++
	statusBefore := strings.ToLower(job.Status)
	job.Status = string(statusRetrying)
	matchedCount, affectedCount, err := persistent.UpdateJob(
		t.ctx, &Filter{JobID: &job.ID}, job.toMap(),
	)
	persistent.Summary().IncrementSummary(ctx, job.TaskName, map[string]interface{}{
		string(job.Status): affectedCount,
		statusBefore:       -matchedCount,
	})
	broadcastAllToSubscribers(t.ctx)
	statusBefore = strings.ToLower(job.Status)

	if defaultOption.debugMode {
		log.Printf("\x1b[35;3mTask Queue Worker: executing task '%s' (job id: %s)\x1b[0m", job.TaskName, job.ID)
	}

	tryRegisterNextJob(ctx, runningTask.taskName)

	trace, ctx := tracer.StartTraceFromHeader(ctx, "TaskQueueWorker", map[string]string{})
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
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
		trace.Finish(tracer.FinishWithError(err))

		matchedCount, affectedCount, _ := persistent.UpdateJob(
			t.ctx,
			&Filter{JobID: &job.ID, Status: candihelper.ToStringPtr(strings.ToUpper(statusBefore))},
			job.toMap(),
			retryHistory,
		)
		if affectedCount == 0 && matchedCount == 0 {
			persistent.SaveJob(t.ctx, &job, retryHistory)
		}
		persistent.Summary().IncrementSummary(ctx, job.TaskName, map[string]interface{}{
			job.Status:   affectedCount,
			statusBefore: -matchedCount,
		})
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

		job.Error = err.Error()
		job.Status = string(statusFailure)

		switch e := err.(type) {
		case *candishared.ErrorRetrier:
			job.ErrorStack = e.StackTrace

			if job.Retries < job.MaxRetry {

				if e.Delay <= 0 {
					e.Delay = defaultInterval
				}

				isRetry = true
				job.Interval = e.Delay.String()
				if e.NewRetryIntervalFunc != nil {
					job.Interval = e.NewRetryIntervalFunc(job.Retries).String()
				}

				// update job arguments if in error retry contains new args payload
				if len(e.NewArgsPayload) > 0 {
					job.Arguments = string(e.NewArgsPayload)
				}

				queue.PushJob(ctx, &job)
				registerJobToWorker(&job, selectedTask.workerIndex)
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
			registerJobToWorker(&nextJob, registeredTask[taskName].workerIndex)
		}

	} else {

		StreamAllJob(ctx, &Filter{
			Page: 1, Limit: 10,
			TaskName: taskName,
			Status:   candihelper.ToStringPtr(string(statusQueueing)),
		}, func(job *Job) {
			queue.PushJob(ctx, job)
			registerJobToWorker(job, registeredTask[job.TaskName].workerIndex)
		})

	}

}
