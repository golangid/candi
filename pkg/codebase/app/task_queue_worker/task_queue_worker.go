package taskqueueworker

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

var (
	registeredJob      map[string]types.WorkerHandlerFunc
	workers            []reflect.SelectCase
	refreshWorkerNotif chan struct{}
	mutex              sync.Mutex
	taskQueue          map[int]*shared.Queue
	workerJobIndex     map[int]string
	workerIndexJob     map[string]int
)

type taskQueueWorker struct {
	service    factory.ServiceFactory
	runningJob int
	shutdown   chan struct{}
	wg         sync.WaitGroup
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	refreshWorkerNotif = make(chan struct{})
	registeredJob = make(map[string]types.WorkerHandlerFunc)
	workerJobIndex = make(map[int]string)
	workerIndexJob = make(map[string]int)
	taskQueue = make(map[int]*shared.Queue)
	return &taskQueueWorker{
		service:  service,
		shutdown: make(chan struct{}),
	}
}

func (c *taskQueueWorker) Serve() {

	// add shutdown channel to first index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.shutdown),
	})
	// add refresh worker channel to second index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(refreshWorkerNotif),
	})

	for _, m := range c.service.GetModules() {
		if h := m.WorkerHandler(types.TaskQueue); h != nil {
			for jobID, handler := range h.MountHandlers() {
				registeredJob[jobID] = handler
				workerIndexJob[jobID] = len(workers)
				workers = append(workers, reflect.SelectCase{Dir: reflect.SelectRecv})
				logger.LogYellow(fmt.Sprintf(`[TASK-QUEUE-WORKER] job_name: %s~%s`, m.Name(), jobID))
			}
		}
	}

	fmt.Printf("\x1b[34;1m⇨ Task queue worker running with %d jobs\x1b[0m\n\n", len(registeredJob))
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

		c.wg.Add(1)
		c.runningJob++
		go func(chosen int) {
			defer c.wg.Done()

			trace := tracer.StartTrace(context.Background(), "TaskQueueWorker")
			defer trace.Finish()
			ctx := trace.Context()

			defer func() {
				if r := recover(); r != nil {
					trace.SetError(fmt.Errorf("%v", r))
				}
				c.runningJob--
				logger.LogGreen(tracer.GetTraceURL(ctx))
			}()

			job := getTaskQueue(chosen)
			job.Retries++

			tags := trace.Tags()
			tags["job_id"] = job.ID
			tags["job_args"] = string(job.Args)
			tags["retries"] = job.Retries
			tags["max_retry"] = job.MaxRetry

			if err := job.HandlerFunc(ctx, job.Args); err != nil {
				switch e := err.(type) {
				case *ErrorRetrier:
					if job.Retries >= job.MaxRetry {
						panic("give up, error: " + e.Error())
					}

					job.interval = time.NewTicker(time.Duration(job.Retries) * e.Delay)
					registerJobInWorker(job)
				default:
					panic(e)
				}
			}
		}(chosen)
	}
}

func (c *taskQueueWorker) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping task queue worker...")
	defer deferFunc()

	if len(registeredJob) == 0 {
		return
	}

	c.shutdown <- struct{}{}

	done := make(chan struct{})
	go func() {
		if c.runningJob != 0 {
			fmt.Printf("\nqueue_worker: waiting %d job... ", c.runningJob)
		}
		c.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		fmt.Print("queue_worker: force shutdown ")
	case <-done:
		fmt.Print("queue_worker: success waiting all task until done ")
	}
}
