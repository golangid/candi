package taskqueueworker

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
)

type taskQueueWorker struct {
	ctx                                 context.Context
	ctxCancelFunc                       func()
	isShutdown                          bool
	ready, shutdown, refreshWorkerNotif chan struct{}
	semaphore                           []chan struct{}
	mutex                               sync.Mutex

	service        factory.ServiceFactory
	wg             sync.WaitGroup
	workerChannels []reflect.SelectCase

	configuration *configurationUsecase
	subscriber    *subscriber
	opt           *option

	registeredTaskWorkerIndex map[string]int
	runningWorkerIndexTask    map[int]*Task
	tasks                     []string

	globalSemaphore chan struct{}
	messagePool     sync.Pool
}

// NewTaskQueueWorker create new task queue worker
func NewTaskQueueWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	e := initEngine(service, opts...)

	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.TaskQueue); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				if _, ok := e.registeredTaskWorkerIndex[handler.Pattern]; ok {
					panic("Task Queue Worker: task \"" + handler.Pattern + "\" has been registered")
				}

				workerIndex := len(e.workerChannels)
				e.registeredTaskWorkerIndex[handler.Pattern] = workerIndex
				e.runningWorkerIndexTask[workerIndex] = &Task{
					handler: handler, moduleName: string(m.Name()),
					taskName: handler.Pattern, workerIndex: workerIndex,
				}
				e.tasks = append(e.tasks, handler.Pattern)
				e.workerChannels = append(e.workerChannels, reflect.SelectCase{Dir: reflect.SelectRecv})
				e.semaphore = append(e.semaphore, make(chan struct{}, 1))

				logger.LogYellow(fmt.Sprintf(`[TASK-QUEUE-WORKER] (task name): %-15s  --> (module): "%s"`, `"`+handler.Pattern+`"`, m.Name()))
			}
		}
	}

	go e.prepare()

	fmt.Printf("\x1b[34;1m⇨ Task Queue Worker running with %d task. Open http://127.0.0.1:%d for dashboard\x1b[0m\n\n",
		len(e.registeredTaskWorkerIndex), e.opt.dashboardPort)

	return e
}

func (t *taskQueueWorker) prepare() {
	if len(t.tasks) == 0 {
		logger.LogYellow("Task Queue Worker: warning, no task provided")
		t.ready <- struct{}{}
		return
	}

	t.opt.locker.Reset(t.getLockKey("*"))
	t.opt.persistent.Summary().DeleteAllSummary(t.ctx, &Filter{ExcludeTaskNameList: t.tasks})
	t.opt.persistent.CleanJob(t.ctx, &Filter{ExcludeTaskNameList: t.tasks})

	// get current pending jobs
	filter := &Filter{
		Page: 1, Limit: 50,
		TaskNameList: t.tasks,
		Sort:         "created_at",
		Statuses:     []string{string(StatusRetrying), string(StatusQueueing)},
	}
	for _, taskName := range t.tasks {
		t.opt.queue.Clear(t.ctx, taskName)
		updated := map[string]interface{}{
			"is_loading": true, "loading_message": "Requeueing...",
		}
		for _, status := range []string{
			StatusRetrying.String(), StatusFailure.String(), StatusSuccess.String(),
			StatusQueueing.String(), StatusStopped.String(),
		} {
			updated[strings.ToLower(status)] = t.opt.persistent.CountAllJob(t.ctx, &Filter{
				TaskName: taskName, Status: &status,
			})
		}
		t.opt.persistent.Summary().UpdateSummary(t.ctx, taskName, updated)
	}
	t.subscriber.broadcastTaskList(t.ctx)

	StreamAllJob(t.ctx, filter, func(job *Job) {
		if t.opt.locker.HasBeenLocked(t.getLockKey(job.ID)) {
			return
		}

		// update to queueing
		if job.Status != string(StatusQueueing) {
			statusBefore := job.Status
			job.Status = string(StatusQueueing)
			matchedCount, affectedCount, err := t.opt.persistent.UpdateJob(t.ctx, &Filter{
				JobID: &job.ID,
			}, map[string]interface{}{
				"status": job.Status,
			})
			if err != nil {
				logger.LogE(err.Error())
				return
			}
			t.opt.persistent.Summary().IncrementSummary(t.ctx, job.TaskName, map[string]int64{
				string(job.Status): affectedCount,
				statusBefore:       -matchedCount,
			})
		}
		t.opt.queue.PushJob(t.ctx, job)
	})

	t.registerInternalTask()
	t.ready <- struct{}{}

	for _, taskName := range t.tasks {
		t.opt.persistent.Summary().UpdateSummary(t.ctx, taskName, map[string]interface{}{
			"is_loading": false, "loading_message": "",
		})
		t.registerNextJob(false, taskName)
	}
	t.subscriber.broadcastTaskList(t.ctx)
}

func (t *taskQueueWorker) Serve() {

	// serve graphql api for communication to dashboard
	go t.serveGraphQLAPI()

	<-t.ready

	// run worker
	for {
		select {
		case <-t.shutdown:
			return
		default:
		}

		chosen, _, ok := reflect.Select(t.workerChannels)
		if !ok {
			logger.LogRed("invalid select worker channels")
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

	if len(t.registeredTaskWorkerIndex) == 0 {
		return
	}

	t.stopAllJob()
	t.shutdown <- struct{}{}
	t.isShutdown = true
	t.opt.locker.Reset(t.getLockKey("*"))

	var runningJob int
	for _, sem := range t.semaphore {
		runningJob += len(sem)
	}
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mTask Queue Worker:\x1b[0m waiting %d job until done...\x1b[0m\n", runningJob)
	}

	t.wg.Wait()
	for _, task := range t.tasks {
		t.opt.queue.Clear(ctx, task)
	}
	t.ctxCancelFunc()
}

func (t *taskQueueWorker) Name() string {
	return string(types.TaskQueue)
}

func (t *taskQueueWorker) registerJobToWorker(job *Job) {
	interval, err := time.ParseDuration(job.Interval)
	if err != nil || interval <= 0 {
		logger.LogRed("invalid interval " + job.Interval)
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	workerIndex := t.registeredTaskWorkerIndex[job.TaskName]
	taskIndex := t.runningWorkerIndexTask[workerIndex]
	taskIndex.activeInterval = time.NewTicker(interval)
	t.workerChannels[workerIndex].Chan = reflect.ValueOf(taskIndex.activeInterval.C)
	t.doRefreshWorker()
}

func (t *taskQueueWorker) stopAllJob() {
	for _, task := range t.runningWorkerIndexTask {
		if task != nil && task.activeInterval != nil {
			task.activeInterval.Stop()
		}
	}
}

func (t *taskQueueWorker) stopAllJobInTask(taskName string) {
	workerIndex, ok := t.registeredTaskWorkerIndex[taskName]
	if !ok {
		return
	}

	if task := t.runningWorkerIndexTask[workerIndex]; task != nil {
		if task.activeInterval != nil {
			task.activeInterval.Stop()
		}
		if task.cancel != nil {
			task.cancel()
		}
	}
}

func (t *taskQueueWorker) doRefreshWorker() {
	t.refreshWorkerNotif <- struct{}{}
}

func (t *taskQueueWorker) releaseMessagePool(eventContext *candishared.EventContext) {
	eventContext.Reset()
	t.messagePool.Put(eventContext)
}
