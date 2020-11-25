package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config/env"
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
	queue                                   QueueStorage
	refreshWorkerNotif, shutdown, semaphore chan struct{}
	mutex                                   sync.Mutex
)

type taskQueueWorker struct {
	service factory.ServiceFactory
	wg      sync.WaitGroup
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	if service.GetDependency().GetRedisPool() == nil {
		panic("Task queue worker require redis for queue storage")
	}

	queue = NewRedisQueue(service.GetDependency().GetRedisPool().WritePool())
	refreshWorkerNotif, shutdown, semaphore = make(chan struct{}), make(chan struct{}, 1), make(chan struct{}, env.BaseEnv().MaxGoroutines)

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

		semaphore <- struct{}{}
		t.wg.Add(1)
		go func(chosen int) {
			defer func() {
				t.wg.Done()
				<-semaphore
			}()

			execJob(chosen)
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
