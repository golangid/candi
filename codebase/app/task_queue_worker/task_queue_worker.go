package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

type taskQueueWorker struct {
	service factory.ServiceFactory
	wg      sync.WaitGroup
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	makeAllGlobalVars(service)

	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.TaskQueue); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				workerIndex := len(workers)
				registeredTask[handler.Pattern] = struct {
					handlerFunc   types.WorkerHandlerFunc
					errorHandlers []types.WorkerErrorHandler
					workerIndex   int
				}{
					handlerFunc: handler.HandlerFunc, workerIndex: workerIndex, errorHandlers: handler.ErrorHandler,
				}
				workerIndexTask[workerIndex] = &struct {
					taskName       string
					activeInterval *time.Ticker
				}{
					taskName: handler.Pattern,
				}
				tasks = append(tasks, handler.Pattern)
				workers = append(workers, reflect.SelectCase{Dir: reflect.SelectRecv})

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
	}()

	fmt.Printf("\x1b[34;1m⇨ Task queue worker running with %d task. Open [::]:%d for dashboard\x1b[0m\n\n",
		len(registeredTask), env.BaseEnv().TaskQueueDashboardPort)

	return &taskQueueWorker{
		service: service,
	}
}

func (t *taskQueueWorker) Serve() {
	// serve graphql api for communication to dashboard
	go serveGraphQLAPI(t)

	// run worker
	for {

		chosen, _, ok := reflect.Select(workers)
		if !ok {
			continue
		}

		// if shutdown channel captured, break loop (no more jobs will run)
		if chosen == 0 {
			break
		}

		// notify for refresh worker
		if chosen == 1 {
			continue
		}

		semaphore <- struct{}{}
		t.wg.Add(1)
		go func(chosen int) {
			defer func() {
				recover()
				t.wg.Done()
				refreshWorkerNotif <- struct{}{}
				<-semaphore
			}()

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
	runningJob := len(semaphore)
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mTask Queue Worker:\x1b[0m waiting %d job until done...\x1b[0m\n", runningJob)
	}

	t.wg.Wait()
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
	taskIndex.activeInterval = nil
	job := queue.PopJob(taskIndex.taskName)
	job, err := repo.findJobByID(job.ID)
	if err != nil {
		return
	}
	if job.Status == string(statusStopped) {
		nextJob := queue.NextJob(taskIndex.taskName)
		if nextJob != nil {
			if jb, err := repo.findJobByID(nextJob.ID); err == nil {
				nextJob = &jb
			}
			registerJobToWorker(nextJob, workerIndex)
		}
		return
	}

	trace := tracer.StartTrace(context.Background(), "TaskQueueWorker")
	defer trace.Finish()
	ctx := trace.Context()

	defer func() {
		if r := recover(); r != nil {
			trace.SetError(fmt.Errorf("%v", r))
		}
		job.FinishedAt = time.Now().Format(time.RFC3339)
		repo.saveJob(job)
		broadcastAllToSubscribers()
		logger.LogGreen("task_queue > trace_url: " + tracer.GetTraceURL(ctx))
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

	nextJob := queue.NextJob(taskIndex.taskName)
	if nextJob != nil {
		if jb, err := repo.findJobByID(nextJob.ID); err == nil {
			nextJob = &jb
		}
		registerJobToWorker(nextJob, workerIndex)
	}

	ctx = context.WithValue(ctx, candishared.ContextKeyTaskQueueRetry, job.Retries)
	if err := registeredTask[job.TaskName].handlerFunc(ctx, []byte(job.Arguments)); err != nil {
		job.Error = err.Error()
		trace.SetError(err)
		switch e := err.(type) {
		case *ErrorRetrier:
			job.Status = string(statusQueueing)
			if job.Retries >= job.MaxRetry {
				fmt.Printf("\x1b[31;1mTaskQueueWorker: GIVE UP: %s\x1b[0m\n", job.TaskName)
				job.Status = string(statusFailure)
				for _, errHandler := range registeredTask[job.TaskName].errorHandlers {
					errHandler(ctx, types.TaskQueue, job.TaskName, []byte(job.Arguments), err)
				}
				return
			}

			delay := e.Delay
			if nextJob != nil && nextJob.Retries == 0 {
				nextJobDelay, _ := time.ParseDuration(nextJob.Interval)
				delay += nextJobDelay
			}

			taskIndex.activeInterval = time.NewTicker(delay)
			workers[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)

			tags["is_retry"] = true

			job.Interval = delay.String()
			queue.PushJob(&job)
		}
	} else {
		job.Status = string(statusSuccess)
	}
}
