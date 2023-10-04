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
	runningTask, ok := t.runningWorkerIndexTask[workerIndex]
	if !ok {
		return
	}

	runningTask.activeInterval.Stop()

	if runningTask.isInternalTask {
		t.execInternalTask(runningTask)
		return
	}

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

			t.wg.Done()
			<-t.semaphore[workerIndex-1]
			task.cancel()

			t.registerNextJob(true, task.taskName)
		}()

		if t.ctx.Err() != nil {
			logger.LogRed("task_queue_worker > ctx root err: " + t.ctx.Err().Error())
			return
		}

		// lock for multiple worker (if running on multiple runtime)
		lockKey := t.getLockKey(runningTask.taskName)
		if t.opt.locker.IsLocked(lockKey) {
			logger.LogI("task_queue_worker > task " + runningTask.taskName + " is locked")
			t.checkForUnlockTask(runningTask.taskName)
			return
		}
		defer t.opt.locker.Unlock(lockKey)

		t.execJob(ctx, task)

	}(workerIndex, runningTask)
}

func (t *taskQueueWorker) execJob(ctx context.Context, runningTask *Task) {
	jobID := t.opt.queue.PopJob(t.ctx, runningTask.taskName)
	if jobID == "" {
		logger.LogI("task_queue_worker > empty queue")
		return
	}

	job, err := t.opt.persistent.FindJobByID(ctx, jobID, nil)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	if job.Status != string(StatusQueueing) {
		logger.LogI("task_queue_worker > skip exec job, job status: " + job.Status + ", job id: " + job.ID)
		return
	}

	selectedHandler := runningTask.handler
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	isContextCanceled, isUpdateArgs, startAt := false, false, time.Now()

	job.Retries++
	statusBefore := strings.ToLower(job.Status)
	job.Status = string(StatusRetrying)
	matchedCount, affectedCount, err := t.opt.persistent.UpdateJob(
		t.ctx, &Filter{JobID: &job.ID}, map[string]interface{}{"status": job.Status},
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
			trace.SetTag("panic", true)
			err = fmt.Errorf("%v", r)
			job.Error = err.Error()
			job.Status = string(StatusFailure)
		}

		job.FinishedAt = time.Now()

		logger.LogGreen("task_queue_worker > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish(tracer.FinishWithError(err))

		incr := map[string]int64{}
		if ok, _ := runningTask.handler.Configs[TaskOptionDeleteJobAfterSuccess].(bool); ok && job.Status == string(StatusSuccess) {
			t.opt.persistent.DeleteJob(t.ctx, job.ID)
			incr = map[string]int64{statusBefore: -1}

		} else {
			updated := map[string]interface{}{
				"retries": job.Retries, "finished_at": job.FinishedAt, "status": job.Status,
				"interval": job.Interval, "error": job.Error, "trace_id": job.TraceID,
				"result": job.Result,
			}
			if isUpdateArgs {
				updated["arguments"] = job.Arguments
			}
			if isContextCanceled {
				updated = map[string]interface{}{
					"error": job.Error, "trace_id": job.TraceID,
				}
			}

			retryHistory := RetryHistory{
				Status: job.Status, Error: job.Error, TraceID: job.TraceID,
				StartAt: startAt, EndAt: job.FinishedAt,
				ErrorStack: job.ErrorStack, Result: job.Result,
			}
			if err != nil {
				retryHistory.Status = StatusFailure.String()
			}
			matchedCount, affectedCount, err := t.opt.persistent.UpdateJob(
				t.ctx, &Filter{JobID: &job.ID}, updated, retryHistory,
			)
			if err != nil {
				updated["error"] = fmt.Sprintf("[Internal Worker Error]: %s", err.Error())
				matchedCount, affectedCount, _ = t.opt.persistent.UpdateJob(
					t.ctx, &Filter{JobID: &job.ID}, updated, retryHistory,
				)
			}
			if !isContextCanceled {
				incr = map[string]int64{
					job.Status: affectedCount, statusBefore: -matchedCount,
				}
			}
		}
		t.opt.persistent.Summary().IncrementSummary(ctx, job.TaskName, incr)
		t.subscriber.broadcastAllToSubscribers(t.ctx)
	}()

	tags := trace.Tags()
	tags["job_id"], tags["task_name"], tags["retries"], tags["max_retry"] = job.ID, job.TaskName, job.Retries, job.MaxRetry
	trace.Log("job_args", job.Arguments)

	job.TraceID = tracer.GetTraceID(ctx)

	eventContext := t.messagePool.Get().(*candishared.EventContext)
	defer t.releaseMessagePool(eventContext)
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(types.TaskQueue))
	eventContext.SetHandlerRoute(job.TaskName)
	eventContext.SetHeader(map[string]string{
		HeaderRetries:         strconv.Itoa(job.Retries),
		HeaderMaxRetries:      strconv.Itoa(job.MaxRetry),
		HeaderInterval:        job.Interval,
		HeaderCurrentProgress: strconv.Itoa(job.CurrentProgress),
		HeaderMaxProgress:     strconv.Itoa(job.MaxProgress),
	})
	eventContext.SetKey(job.ID)
	eventContext.WriteString(job.Arguments)

	if len(selectedHandler.HandlerFuncs) == 0 {
		job.Error = "No handler found for exec this job"
		job.Status = string(StatusFailure)
		return
	}

	errChan := make(chan error)
	go func(e *candishared.EventContext) {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic: %v", r)
			}
			close(errChan)
		}()

		errChan <- selectedHandler.HandlerFuncs[0](e)
	}(eventContext)

	select {
	case <-ctx.Done():
		job.Error = "Job has been stopped when running (context canceled)"
		job.Status = string(StatusStopped)
		isContextCanceled = true
		return

	case err = <-errChan:
		job.Status = string(StatusSuccess)
		job.Error = ""
		if respBuff := eventContext.GetResponse(); respBuff != nil {
			job.Result = respBuff.String()
		}
		if err != nil {
			eventContext.SetError(err)

			job.Error = err.Error()
			job.Status = string(StatusFailure)

			switch e := err.(type) {
			case *candishared.ErrorRetrier:
				job.ErrorStack = e.StackTrace

				if job.Retries < job.MaxRetry {
					if e.Delay <= 0 {
						e.Delay = defaultInterval
					}

					job.Status = string(StatusQueueing)
					trace.SetTag("is_retry", true)
					job.Interval = e.Delay.String()
					if e.NewRetryIntervalFunc != nil {
						job.Interval = e.NewRetryIntervalFunc(job.Retries).String()
					}

					// update job arguments if in error retry contains new args payload
					if len(e.NewArgsPayload) > 0 {
						job.Arguments = string(e.NewArgsPayload)
						isUpdateArgs = true
					}

					t.opt.queue.PushJob(ctx, &job)
					return
				}

				logger.LogRed("TaskQueueWorker: Still error for task '" + job.TaskName + "' (job id: " + job.ID + ")")
			}
		}

		for _, h := range selectedHandler.HandlerFuncs[1:] {
			h(eventContext)
		}
		return

	}
}

func (t *taskQueueWorker) getLockKey(jobID string) string {
	return fmt.Sprintf("%s:task-queue-worker-lock:%s", t.service.Name(), jobID)
}

func (t *taskQueueWorker) checkForUnlockTask(taskName string) {
	if count := t.opt.persistent.CountAllJob(t.ctx, &Filter{
		TaskName: taskName, Status: candihelper.ToStringPtr(StatusRetrying.String()),
	}); count == 0 {
		t.opt.locker.Unlock(t.getLockKey(taskName))
	}
}

func (t *taskQueueWorker) registerNextJob(withStream bool, taskName string) {
	nextJobID := t.opt.queue.NextJob(t.ctx, taskName)
	if nextJobID != "" {
		if nextJob, err := t.opt.persistent.FindJobByID(t.ctx, nextJobID, nil); err == nil {
			t.registerJobToWorker(&nextJob)
			return
		}

		t.opt.queue.PopJob(t.ctx, taskName) // remove unused job (job not found maybe if not save job after success)
		t.registerNextJob(false, taskName)

	} else if withStream {
		StreamAllJob(t.ctx, &Filter{
			TaskName: taskName,
			Sort:     "created_at",
			Status:   candihelper.ToStringPtr(string(StatusQueueing)),
		}, func(job *Job) {
			t.opt.queue.PushJob(t.ctx, job)
		})
		t.registerNextJob(false, taskName)
	}
}
