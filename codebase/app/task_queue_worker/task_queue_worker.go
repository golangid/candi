package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"math"
	"reflect"
	"sync"

	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
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
			filter := Filter{
				Page: 1, Limit: 10,
				TaskNameList: tasks,
				Status:       []string{string(statusRetrying), string(statusQueueing)},
			}
			count := persistent.CountAllJob(workerInstance.ctx, filter)
			totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
			for filter.Page <= totalPages {
				for _, job := range persistent.FindAllJob(workerInstance.ctx, filter) {
					queue.PushJob(workerInstance.ctx, &job)
					registerJobToWorker(&job, registeredTask[job.TaskName].workerIndex)
				}
				filter.Page++
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

	for _, task := range tasks {
		queue.Clear(ctx, task)
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

	newCtx := context.Background()
	select {
	case <-ctx.Done():
		persistent.UpdateJob(newCtx, Filter{
			Status: []string{string(statusRetrying)},
		}, map[string]interface{}{
			"status": statusQueueing,
		})
		broadcastAllToSubscribers(newCtx)
	case <-done:
		broadcastAllToSubscribers(newCtx)
		t.ctxCancelFunc()
	}
}

func (t *taskQueueWorker) Name() string {
	return string(types.TaskQueue)
}
