package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
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

	runningTask, ok := t.runningWorkerIndexTask[workerIndex]
	if !ok {
		return
	}

	if runningTask.isInternalTask {
		t.execInternalTask(runningTask.internalTaskName)
		return
	}

	runningTask.activeInterval.Stop()

	t.semaphore[workerIndex-1] <- struct{}{}
	if t.isShutdown {
		logger.LogRed("worker has been shutdown")
		return
	}

	t.wg.Add(1)
	go func(workerIndex int, task *Task) {
		var ctx context.Context
		ctx, task.cancel = context.WithCancel(t.ctx)

		defer func() {
			if r := recover(); r != nil {
				logger.LogRed(fmt.Sprintf("task_queue_worker > panic: %v", r))
			}
			t.registerNextJob(task.taskName)
			t.wg.Done()
			if ctx.Err() == nil {
				<-t.semaphore[workerIndex-1]
				task.cancel()
			}
			t.refreshWorkerNotif <- struct{}{}
		}()

		if t.ctx.Err() != nil {
			logger.LogRed("task_queue_worker > ctx root err: " + t.ctx.Err().Error())
			return
		}

		t.execJob(ctx, task)

	}(workerIndex, runningTask)
}

func (t *taskQueueWorker) execJob(ctx context.Context, runningTask *Task) {
	jobID := t.opt.queue.PopJob(t.ctx, runningTask.taskName)
	if jobID == "" {
		logger.LogYellow("task_queue_worker > empty queue")
		return
	}

	// lock for multiple worker (if running on multiple pods/instance)
	if t.opt.locker.IsLocked(t.getLockKey(jobID)) {
		logger.LogYellow("task_queue_worker > job " + jobID + " is locked")
		return
	}
	defer t.opt.locker.Unlock(t.getLockKey(jobID))

	selectedTask := t.registeredTask[runningTask.taskName]

	job, err := t.opt.persistent.FindJobByID(ctx, jobID, nil)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	if job.Status != string(statusQueueing) {
		logger.LogYellow("task_queue_worker > skip exec job, job status: " + job.Status)
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
	matchedCount, affectedCount, err := t.opt.persistent.UpdateJob(
		t.ctx, &Filter{JobID: &job.ID}, job.toMap(),
	)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	t.opt.persistent.Summary().IncrementSummary(ctx, job.TaskName, map[string]int64{
		string(job.Status): affectedCount,
		statusBefore:       -matchedCount,
	})
	t.subscriber.broadcastAllToSubscribers(t.ctx)
	statusBefore = strings.ToLower(job.Status)

	if t.opt.debugMode {
		log.Printf("\x1b[35;3mTask Queue Worker: executing task '%s' (job id: %s)\x1b[0m", job.TaskName, job.ID)
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "TaskQueueWorker", map[string]string{})
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			job.Error = err.Error()
			job.Status = string(statusFailure)
			trace.Log("stacktrace", string(debug.Stack()))
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

		matchedCount, affectedCount, _ := t.opt.persistent.UpdateJob(
			t.ctx,
			&Filter{JobID: &job.ID, Status: candihelper.ToStringPtr(strings.ToUpper(statusBefore))},
			job.toMap(),
			retryHistory,
		)
		if affectedCount == 0 && matchedCount == 0 {
			t.opt.persistent.SaveJob(t.ctx, &job, retryHistory)
		}
		t.opt.persistent.Summary().IncrementSummary(ctx, job.TaskName, map[string]int64{
			job.Status:   affectedCount,
			statusBefore: -matchedCount,
		})
		t.subscriber.broadcastAllToSubscribers(t.ctx)
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
		logger.LogE(ctx.Err().Error())
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

				t.opt.queue.PushJob(ctx, &job)
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

func (t *taskQueueWorker) registerNextJob(taskName string) {

	nextJobID := t.opt.queue.NextJob(t.ctx, taskName)
	if nextJobID != "" {

		if nextJob, err := t.opt.persistent.FindJobByID(t.ctx, nextJobID, nil); err == nil {
			t.registerJobToWorker(&nextJob, t.registeredTask[taskName].workerIndex)
		}

	} else {

		StreamAllJob(t.ctx, &Filter{
			Page: 1, Limit: 10,
			TaskName: taskName,
			Sort:     "created_at",
			Status:   candihelper.ToStringPtr(string(statusQueueing)),
		}, func(job *Job) {
			if n := t.opt.queue.PushJob(t.ctx, job); n <= 1 {
				t.registerJobToWorker(job, t.registeredTask[job.TaskName].workerIndex)
			}
		})

	}

}

func (t *taskQueueWorker) execInternalTask(internalTaskName string) {

	logger.LogIf("running internal task: %s", internalTaskName)

	switch internalTaskName {
	case configurationRetentionAgeKey:

		cfg, _ := t.opt.persistent.GetConfiguration(configurationRetentionAgeKey)
		if !cfg.IsActive {
			return
		}
		dateDuration, err := time.ParseDuration(cfg.Value)
		if err != nil || dateDuration <= 0 {
			return
		}

		beforeCreatedAt := time.Now().Add(-dateDuration)
		affectedStatus := []string{string(statusSuccess), string(statusFailure), string(statusStopped)}
		for _, task := range t.tasks {
			incrQuery := map[string]int64{}
			for _, status := range affectedStatus {
				countAffected := t.opt.persistent.CleanJob(t.ctx,
					&Filter{
						TaskName: task, Status: &status, BeforeCreatedAt: &beforeCreatedAt,
					},
				)
				incrQuery[strings.ToLower(status)] -= countAffected
			}
			t.opt.persistent.Summary().IncrementSummary(t.ctx, task, incrQuery)
		}
		t.subscriber.broadcastAllToSubscribers(t.ctx)
	}

}
