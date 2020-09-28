package taskqueueworker

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/logger"
)

var (
	registeredTask map[string]struct {
		handlerFunc   types.WorkerHandlerFunc
		errorHandlers []types.WorkerErrorHandler
		workerIndex   int
	}

	workers         []reflect.SelectCase
	workerIndexTask map[int]*struct {
		taskName       string
		activeInterval *time.Ticker
	}
	queue                        QueueStorage
	refreshWorkerNotif, shutdown chan struct{}
	mutex                        sync.Mutex
)

type taskQueueWorker struct {
	service    factory.ServiceFactory
	runningJob int
	wg         sync.WaitGroup
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	if service.GetDependency().GetRedisPool() == nil {
		panic("Task queue worker require redis for queue storage")
	}

	queue = NewRedisQueue(service.GetDependency().GetRedisPool().WritePool())
	refreshWorkerNotif, shutdown = make(chan struct{}), make(chan struct{})
	registeredTask = make(map[string]struct {
		handlerFunc   types.WorkerHandlerFunc
		errorHandlers []types.WorkerErrorHandler
		workerIndex   int
	})
	workerIndexTask = make(map[int]*struct {
		taskName       string
		activeInterval *time.Ticker
	})

	// add shutdown channel to first index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(shutdown),
	})
	// add refresh worker channel to second index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(refreshWorkerNotif),
	})

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
				workers = append(workers, reflect.SelectCase{Dir: reflect.SelectRecv})

				logger.LogYellow(fmt.Sprintf(`[TASK-QUEUE-WORKER] Task name: %s`, handler.Pattern))
			}
		}
	}

	// get current queue
	for taskName, registered := range registeredTask {
		for _, job := range queue.GetAllJobs(taskName) {
			registerJobToWorker(job, registered.workerIndex)
		}
	}

	fmt.Printf("\x1b[34;1m⇨ Task queue worker running with %d task\x1b[0m\n\n", len(registeredTask))

	return &taskQueueWorker{
		service: service,
	}
}

func (t *taskQueueWorker) Serve() {

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

		t.wg.Add(1)
		t.runningJob++
		go func(chosen int) {
			defer t.wg.Done()
			t.runningJob--

			execJob(chosen)
		}(chosen)
	}
}

func (t *taskQueueWorker) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping task queue worker...")
	defer deferFunc()

	if len(registeredTask) == 0 {
		return
	}

	shutdown <- struct{}{}

	done := make(chan struct{})
	go func() {
		if t.runningJob != 0 {
			fmt.Printf("\nqueue_worker: waiting %d job... ", t.runningJob)
		}
		t.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		fmt.Print("queue_worker: force shutdown ")
	case <-done:
		fmt.Print("queue_worker: success waiting all task until done ")
	}
}
