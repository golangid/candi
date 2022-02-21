package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

type taskQueueWorker struct {
	ctx           context.Context
	ctxCancelFunc func()
	isShutdown    bool

	service factory.ServiceFactory
	wg      sync.WaitGroup
}

// NewTaskQueueWorker create new task queue worker
func NewTaskQueueWorker(service factory.ServiceFactory, q QueueStorage, perst Persistent, opts ...OptionFunc) factory.AppServerFactory {
	makeAllGlobalVars(q, perst, opts...)

	workerInstance := &taskQueueWorker{
		service: service,
	}
	workerInstance.ctx, workerInstance.ctxCancelFunc = context.WithCancel(context.Background())
	defaultOption.locker.Reset(fmt.Sprintf("%s:task-queue-worker-lock:*", service.Name()))

	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.TaskQueue); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				if _, ok := registeredTask[handler.Pattern]; ok {
					panic("Task Queue Worker: task " + handler.Pattern + " has been registered")
				}

				workerIndex := len(workers)
				registeredTask[handler.Pattern] = struct {
					handler     types.WorkerHandler
					workerIndex int
					moduleName  string
				}{
					handler: handler, workerIndex: workerIndex, moduleName: string(m.Name()),
				}
				workerIndexTask[workerIndex] = &Task{
					taskName: handler.Pattern,
				}
				tasks = append(tasks, handler.Pattern)
				workers = append(workers, reflect.SelectCase{Dir: reflect.SelectRecv})
				semaphore = append(semaphore, make(chan struct{}, 1))

				logger.LogYellow(fmt.Sprintf(`[TASK-QUEUE-WORKER] (task name): %-15s  --> (module): "%s"`, `"`+handler.Pattern+`"`, m.Name()))
			}
		}
	}

	if len(tasks) == 0 {
		logger.LogYellow("Task Queue Worker: warning, no task provided")

	} else {

		go func() {
			for _, taskName := range tasks {
				queue.Clear(workerInstance.ctx, taskName)
			}
			// get current pending jobs
			pageNumber := 1
			filter := Filter{
				TaskNameList: tasks,
				Status:       []string{string(statusRetrying), string(statusQueueing)},
				Limit:        100,
			}
			count := persistent.CountAllJob(workerInstance.ctx, filter)
			totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
			for pageNumber <= totalPages {
				filter.Page = pageNumber
				pendingJobs := persistent.FindAllJob(workerInstance.ctx, filter)
				for _, job := range pendingJobs {
					queue.PushJob(workerInstance.ctx, &job)
					registerJobToWorker(&job, registeredTask[job.TaskName].workerIndex)
				}
				pageNumber++
			}
			refreshWorkerNotif <- struct{}{}
		}()
	}

	fmt.Printf("\x1b[34;1m⇨ Task Queue Worker running with %d task. Open [::]:%d for dashboard\x1b[0m\n\n",
		len(registeredTask), defaultOption.dashboardPort)

	return workerInstance
}

func (t *taskQueueWorker) Serve() {
	// serve graphql api for communication to dashboard
	go serveGraphQLAPI(t)

	// run worker
	for {
		select {
		case <-shutdown:
			return
		default:
		}

		chosen, _, ok := reflect.Select(workers)
		if !ok {
			continue
		}

		// notify for refresh worker
		if chosen == 0 {
			continue
		}

		go t.triggerTask(chosen)
	}
}

func (t *taskQueueWorker) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping Task Queue Worker:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	if len(registeredTask) == 0 {
		return
	}

	stopAllJob()
	shutdown <- struct{}{}
	t.isShutdown = true
	var runningJob int
	for _, sem := range semaphore {
		runningJob += len(sem)
	}
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mTask Queue Worker:\x1b[0m waiting %d job until done...\x1b[0m\n", runningJob)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		t.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		persistent.UpdateAllStatus(t.ctx, "", []JobStatusEnum{statusRetrying}, statusQueueing)
		broadcastAllToSubscribers(t.ctx)
	case <-done:
		broadcastAllToSubscribers(t.ctx)
		t.ctxCancelFunc()
	}
}

func (t *taskQueueWorker) Name() string {
	return string(types.TaskQueue)
}

func (t *taskQueueWorker) triggerTask(workerIndex int) {
	semaphore[workerIndex-1] <- struct{}{}
	if t.isShutdown {
		return
	}

	t.wg.Add(1)
	go func(workerIndex int) {
		defer func() {
			recover()
			t.wg.Done()
			<-semaphore[workerIndex-1]
			refreshWorkerNotif <- struct{}{}
		}()

		if t.ctx.Err() != nil {
			logger.LogRed("task_queue_worker > ctx root err: " + t.ctx.Err().Error())
			return
		}

		runningTask, ok := workerIndexTask[workerIndex]
		if !ok {
			return
		}
		runningTask.activeInterval.Stop()
		t.execJob(runningTask)

	}(workerIndex)

	refreshWorkerNotif <- struct{}{}
}

func (t *taskQueueWorker) execJob(runningTask *Task) {
	jobID := queue.PopJob(t.ctx, runningTask.taskName)
	if jobID == "" {
		return
	}

	// lock for multiple worker (if running on multiple pods/instance)
	if defaultOption.locker.IsLocked(t.getLockKey(jobID)) {
		return
	}
	defer defaultOption.locker.Unlock(t.getLockKey(jobID))

	var ctx context.Context
	ctx, runningTask.cancel = context.WithCancel(t.ctx)
	defer runningTask.cancel()

	selectedTask := registeredTask[runningTask.taskName]

	job, err := persistent.FindJobByID(ctx, jobID, "retry_histories")
	if err != nil || job.Status == string(statusStopped) {
		nextJobID := queue.NextJob(ctx, runningTask.taskName)
		if nextJobID != "" {
			if nextJob, err := persistent.FindJobByID(ctx, nextJobID); err == nil {
				registerJobToWorker(nextJob, selectedTask.workerIndex)
			}
		}
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
	broadcastAllToSubscribers(ctx)

	if defaultOption.debugMode {
		log.Printf("\x1b[35;3mTask Queue Worker: executing task '%s'\x1b[0m", job.TaskName)
	}

	nextJobID := queue.NextJob(ctx, runningTask.taskName)
	if nextJobID != "" {
		if nextJob, err := persistent.FindJobByID(ctx, nextJobID); err == nil {
			registerJobToWorker(nextJob, selectedTask.workerIndex)
		}
	}

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

		persistent.SaveJob(ctx, job, retryHistory)
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
		job.Error = "Job has been stopped when running."
		if ctx.Err() != nil {
			job.Error += " Error: " + ctx.Err().Error()
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
