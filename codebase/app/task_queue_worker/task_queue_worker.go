package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
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
func NewTaskQueueWorker(service factory.ServiceFactory, q QueueStorage, db *mongo.Database, opts ...OptionFunc) factory.AppServerFactory {
	makeAllGlobalVars(q, db, opts...)

	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.TaskQueue); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
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
	}

	go func() {
		// get current pending jobs
		pendingJobs := repo.findAllPendingJob()
		for taskName, registered := range registeredTask {
			queue.Clear(taskName)
			for _, job := range pendingJobs {
				if job.TaskName == taskName {
					queue.PushJob(&job)
					registerJobToWorker(&job, registered.workerIndex)
				}
			}
		}
		refreshWorkerNotif <- struct{}{}
	}()

	fmt.Printf("\x1b[34;1m⇨ Task Queue Worker running with %d task. Open [::]:%d for dashboard\x1b[0m\n\n",
		len(registeredTask), env.BaseEnv().TaskQueueDashboardPort)

	workerInstance := &taskQueueWorker{
		service: service,
	}
	workerInstance.ctx, workerInstance.ctxCancelFunc = context.WithCancel(context.Background())
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
	log.Println("\x1b[33;1mStopping Task Queue Worker...\x1b[0m")
	defer func() { log.Println("\x1b[33;1mStopping Task Queue Worker:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m") }()

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
		repo.pauseAllRunningJob()
	case <-done:
		t.ctxCancelFunc()
	}
	broadcastAllToSubscribers()
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
	job, err := repo.findJobByID(jobID)
	if err != nil {
		return
	}
	if job.Status == string(statusStopped) {
		nextJobID := queue.NextJob(taskIndex.taskName)
		if nextJobID != "" {
			if nextJob, err := repo.findJobByID(nextJobID); err == nil {
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
		repo.saveJob(job)
		broadcastAllToSubscribers()
		logger.LogGreen("task_queue > trace_url: " + tracer.GetTraceURL(ctx))
		trace.Finish()
	}()

	job.Retries++
	job.Status = string(statusRetrying)
	repo.saveJob(job)
	broadcastAllToSubscribers()

	job.TraceID = tracer.GetTraceID(ctx)

	if env.BaseEnv().DebugMode {
		log.Printf("\x1b[35;3mTask Queue Worker: executing task '%s'\x1b[0m", job.TaskName)
	}

	tags := trace.Tags()
	tags["job_id"], tags["task_name"], tags["retries"], tags["max_retry"] = job.ID, job.TaskName, job.Retries, job.MaxRetry
	tracer.Log(ctx, "job_args", job.Arguments)

	nextJobID := queue.NextJob(taskIndex.taskName)
	if nextJobID != "" {
		if nextJob, err := repo.findJobByID(nextJobID); err == nil {
			registerJobToWorker(nextJob, workerIndex)
		}
	}

	ctx = context.WithValue(ctx, candishared.ContextKeyTaskQueueRetry, job.Retries)
	if err := selectedHandler.HandlerFunc(ctx, []byte(job.Arguments)); err != nil {
		job.Error = err.Error()
		job.Status = string(statusFailure)
		trace.SetError(err)

		switch e := err.(type) {
		case *candishared.ErrorRetrier:
			if job.Retries >= job.MaxRetry {
				logger.LogRed("TaskQueueWorker: GIVE UP: " + job.TaskName)
				for _, errHandler := range selectedHandler.ErrorHandler {
					errHandler(ctx, types.TaskQueue, job.TaskName, []byte(job.Arguments), err)
				}
				return
			}
			tags["is_retry"] = true

			job.Status = string(statusQueueing)
			job.Interval = e.Delay.String()

			registerJobToWorker(job, workerIndex)
			queue.PushJob(job)
		}
	} else {
		job.Status = string(statusSuccess)
	}
}
