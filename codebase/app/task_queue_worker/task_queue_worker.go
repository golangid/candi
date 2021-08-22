package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"math"
	"reflect"
	"sync"
	"time"

	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
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
				}{
					handler: handler, workerIndex: workerIndex,
				}
				workerIndexTask[workerIndex] = &struct {
					taskName       string
					activeInterval *time.Ticker
				}{
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
				queue.Clear(taskName)
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
					queue.PushJob(&job)
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

		semaphore[chosen-1] <- struct{}{}
		if t.isShutdown {
			return
		}

		t.wg.Add(1)
		go func(chosen int) {
			defer func() {
				recover()
				t.wg.Done()
				<-semaphore[chosen-1]
				refreshWorkerNotif <- struct{}{}
			}()

			if t.ctx.Err() != nil {
				logger.LogRed("task_queue_worker > ctx root err: " + t.ctx.Err().Error())
				return
			}
			t.execJob(chosen)
		}(chosen)
	}
}

func (t *taskQueueWorker) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping Task Queue Worker:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	if len(registeredTask) == 0 {
		return
	}

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

func (t *taskQueueWorker) execJob(workerIndex int) {
	taskIndex, ok := workerIndexTask[workerIndex]
	if !ok {
		return
	}

	taskIndex.activeInterval.Stop()
	jobID := queue.PopJob(taskIndex.taskName)
	if jobID == "" {
		return
	}
	job, err := persistent.FindJobByID(t.ctx, jobID)
	if err != nil {
		return
	}
	if job.Status == string(statusStopped) {
		nextJobID := queue.NextJob(taskIndex.taskName)
		if nextJobID != "" {
			if nextJob, err := persistent.FindJobByID(t.ctx, nextJobID); err == nil {
				registerJobToWorker(nextJob, workerIndex)
			}
		}
		return
	}

	ctx := t.ctx
	selectedHandler := registeredTask[job.TaskName].handler
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	trace, ctx := tracer.StartTraceWithContext(ctx, "TaskQueueWorker")
	defer func() {
		if r := recover(); r != nil {
			trace.SetError(fmt.Errorf("%v", r))
		}
		job.FinishedAt = time.Now()
		persistent.SaveJob(t.ctx, job)
		broadcastAllToSubscribers(t.ctx)
		logger.LogGreen("task_queue > trace_url: " + tracer.GetTraceURL(ctx))
		trace.Finish()
	}()

	job.Retries++
	job.Status = string(statusRetrying)
	persistent.SaveJob(t.ctx, job)
	broadcastAllToSubscribers(t.ctx)

	job.TraceID = tracer.GetTraceID(ctx)

	if defaultOption.debugMode {
		log.Printf("\x1b[35;3mTask Queue Worker: executing task '%s'\x1b[0m", job.TaskName)
	}

	tags := trace.Tags()
	tags["job_id"], tags["task_name"], tags["retries"], tags["max_retry"] = job.ID, job.TaskName, job.Retries, job.MaxRetry
	tracer.Log(ctx, "job_args", job.Arguments)

	nextJobID := queue.NextJob(taskIndex.taskName)
	if nextJobID != "" {
		if nextJob, err := persistent.FindJobByID(t.ctx, nextJobID); err == nil {
			registerJobToWorker(nextJob, workerIndex)
		}
	}

	message := []byte(job.Arguments)
	ctx = context.WithValue(ctx, candishared.ContextKeyTaskQueueRetry, job.Retries)
	if err := selectedHandler.HandlerFunc(ctx, message); err != nil {
		job.Error = err.Error()
		job.Status = string(statusFailure)
		trace.SetError(err)

		switch e := err.(type) {
		case *candishared.ErrorRetrier:
			if job.Retries >= job.MaxRetry {
				logger.LogRed("TaskQueueWorker: GIVE UP: " + job.TaskName)
				if selectedHandler.ErrorHandler != nil {
					selectedHandler.ErrorHandler(ctx, types.TaskQueue, job.TaskName, message, err)
				}
				return
			}
			tags["is_retry"] = true

			job.Status = string(statusQueueing)
			job.Interval = e.Delay.String()

			registerJobToWorker(job, workerIndex)
			queue.PushJob(job)

		default:
			if selectedHandler.ErrorHandler != nil {
				selectedHandler.ErrorHandler(ctx, types.TaskQueue, job.TaskName, []byte(job.Arguments), err)
			}
		}
	} else {
		job.Status = string(statusSuccess)
	}
}
