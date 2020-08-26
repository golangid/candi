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
)

var (
	registeredTask map[string]struct {
		handlerFunc types.WorkerHandlerFunc
		workerIndex int
	}

	workers         []reflect.SelectCase
	workerIndexTask map[int]*struct {
		taskID         string
		activeInterval *time.Ticker
	}
	queue              QueueStorage
	refreshWorkerNotif chan struct{}
	mutex              sync.Mutex
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
	registeredTask = make(map[string]struct {
		handlerFunc types.WorkerHandlerFunc
		workerIndex int
	})
	workerIndexTask = make(map[int]*struct {
		taskID         string
		activeInterval *time.Ticker
	})
	queue = NewRedisQueue(service.GetDependency().GetRedisPool().WritePool())

	return &taskQueueWorker{
		service:  service,
		shutdown: make(chan struct{}),
	}
}

func (t *taskQueueWorker) Serve() {

	// add shutdown channel to first index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(t.shutdown),
	})
	// add refresh worker channel to second index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(refreshWorkerNotif),
	})

	for _, m := range t.service.GetModules() {
		if h := m.WorkerHandler(types.TaskQueue); h != nil {
			for taskID, handlerFunc := range h.MountHandlers() {
				workerIndex := len(workers)
				registeredTask[taskID] = struct {
					handlerFunc types.WorkerHandlerFunc
					workerIndex int
				}{
					handlerFunc: handlerFunc, workerIndex: workerIndex,
				}
				workerIndexTask[workerIndex] = &struct {
					taskID         string
					activeInterval *time.Ticker
				}{
					taskID: taskID,
				}

				workers = append(workers, reflect.SelectCase{Dir: reflect.SelectRecv})

				logger.LogYellow(fmt.Sprintf(`[TASK-QUEUE-WORKER] Task name: %s`, taskID))
			}
		}
	}

	// get current queue
	for taskID, registered := range registeredTask {
		for _, job := range queue.GetAllJobs(taskID) {
			registerJobToWorker(job, registered.workerIndex)
		}
	}

	fmt.Printf("\x1b[34;1m⇨ Task queue worker running with %d task\x1b[0m\n\n", len(registeredTask))
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

	t.shutdown <- struct{}{}

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
