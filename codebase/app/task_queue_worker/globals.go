package taskqueueworker

import (
	"bytes"
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
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
	if redisPool := service.GetDependency().GetRedisPool(); redisPool != nil {
		opt.locker = candiutils.NewRedisLocker(redisPool.WritePool())
	} else {
		opt.locker = &candiutils.NoopLocker{}
	}
	opt.secondaryPersistent = &noopPersistent{}
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
		if mongoDB := service.GetDependency().GetMongoDatabase(); mongoDB != nil {
			opt.persistent = NewMongoPersistent(mongoDB.WriteDB())
		} else if sqlDB := service.GetDependency().GetSQLDatabase(); sqlDB != nil {
			opt.persistent = NewSQLPersistent(sqlDB.WriteDB())
		} else {
			opt.persistent = NewNoopPersistent()
		}
	}

	if opt.queue == nil {
		if redisPool := service.GetDependency().GetRedisPool(); redisPool != nil {
			opt.queue = NewRedisQueue(redisPool.WritePool())
		} else {
			opt.queue = NewInMemQueue()
		}
	}

	engine = &taskQueueWorker{
		service:                   service,
		ready:                     make(chan struct{}),
		shutdown:                  make(chan struct{}, 1),
		refreshWorkerNotif:        make(chan struct{}, 1),
		opt:                       &opt,
		configuration:             initConfiguration(&opt),
		registeredTaskWorkerIndex: make(map[string]int),
		runningWorkerIndexTask:    make(map[int]*Task),
		globalSemaphore:           make(chan struct{}, env.BaseEnv().MaxGoroutines),
		messagePool: sync.Pool{
			New: func() interface{} {
				return candishared.NewEventContextWithResult(
					bytes.NewBuffer(make([]byte, 0, 256)),
					bytes.NewBuffer(make([]byte, 0, 256)),
				)
			},
		},
	}
	engine.subscriber = initSubscriber(engine.configuration, &opt)
	engine.ctx, engine.ctxCancelFunc = context.WithCancel(context.Background())

	// add refresh worker channel to first index
	engine.workerChannels = append(engine.workerChannels, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(engine.refreshWorkerNotif),
	})

	return engine
}
