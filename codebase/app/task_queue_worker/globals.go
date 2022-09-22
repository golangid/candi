package taskqueueworker

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
)

var (
	// externalWorkerHost setting worker host for add job, if not empty default using http request when add job
	externalWorkerHost string

	// core engine
	engine *taskQueueWorker
)

func initEngine(service factory.ServiceFactory, opts ...OptionFunc) *taskQueueWorker {

	var opt option
	// set default value
	opt.maxClientSubscriber = 5
	opt.autoRemoveClientInterval = 30 * time.Minute
	opt.dashboardPort = 8080
	opt.debugMode = true
	opt.locker = &candiutils.NoopLocker{}
	opt.dashboardBanner = `    _________    _   ______  ____
   / ____/   |  / | / / __ \/  _/
  / /   / /| | /  |/ / / / // /  
 / /___/ ___ |/ /|  / /_/ // /   
 \____/_/  |_/_/ |_/_____/___/   `

	//  override option value
	for _, optFunc := range opts {
		optFunc(&opt)
	}

	// set default persistent & queue if not defined
	if opt.persistent == nil {
		if service.GetDependency().GetMongoDatabase() != nil {
			opt.persistent = NewMongoPersistent(service.GetDependency().GetMongoDatabase().WriteDB())
		} else if service.GetDependency().GetSQLDatabase() != nil {
			opt.persistent = NewSQLPersistent(service.GetDependency().GetSQLDatabase().WriteDB())
		} else {
			opt.persistent = NewNoopPersistent()
		}
	}

	if opt.queue == nil {
		if service.GetDependency().GetRedisPool() != nil {
			opt.queue = NewRedisQueue(service.GetDependency().GetRedisPool().WritePool())
		} else {
			opt.queue = NewInMemQueue()
		}
	}

	engine = &taskQueueWorker{
		service:            service,
		ready:              make(chan struct{}),
		shutdown:           make(chan struct{}, 1),
		refreshWorkerNotif: make(chan struct{}),
		opt:                &opt,
		configuration:      initConfiguration(&opt),
		registeredTask: make(map[string]struct {
			handler     types.WorkerHandler
			workerIndex int
			moduleName  string
		}),
		runningWorkerIndexTask: make(map[int]*Task),
	}
	engine.subscriber = initSubscriber(engine.configuration, &opt)
	engine.ctx, engine.ctxCancelFunc = context.WithCancel(context.Background())
	engine.opt.locker.Reset(fmt.Sprintf("%s:task-queue-worker-lock:*", service.Name()))

	// add refresh worker channel to first index
	engine.workerChannels = append(engine.workerChannels, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(engine.refreshWorkerNotif),
	})

	return engine
}
