package taskqueueworker

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/config/env"
)

type (
	// Task model
	Task struct {
		cancel         context.CancelFunc
		taskName       string
		activeInterval *time.Ticker
	}

	// Filter type
	Filter struct {
		Page, Limit         int
		Sort                string
		TaskName            string
		TaskNameList        []string
		ExcludeTaskNameList []string
		Search, JobID       *string
		Status              *string
		Statuses            []string
		ExcludeStatus       []string
		ShowAll             bool
		ShowHistories       *bool
		StartDate, EndDate  time.Time
		Count               int
	}

	// ClientSubscriber model
	ClientSubscriber struct {
		ClientID      string
		SubscribeList struct {
			TaskDashboard bool
			JobDetailID   string
			JobList       *Filter
		}
	}

	clientTaskDashboardSubscriber struct {
		c      chan TaskListResolver
		filter *Filter
	}
	clientTaskJobListSubscriber struct {
		c             chan JobListResolver
		skipBroadcast bool
		filter        *Filter
	}
	clientJobDetailSubscriber struct {
		c      chan JobResolver
		filter *Filter
	}

	// JobStatusEnum enum status
	JobStatusEnum string
)

func (s *clientTaskDashboardSubscriber) writeDataToChannel(data TaskListResolver) {
	defer func() { recover() }()
	s.c <- data
}
func (s *clientTaskJobListSubscriber) writeDataToChannel(data JobListResolver) {
	defer func() { recover() }()
	s.c <- data
}
func (s *clientJobDetailSubscriber) writeDataToChannel(data JobResolver) {
	defer func() { recover() }()
	s.c <- data
}

// CalculateOffset method
func (f *Filter) CalculateOffset() int {
	return (f.Page - 1) * f.Limit
}

// String method
func (j JobStatusEnum) String() string {
	return string(j)
}

const (
	statusRetrying JobStatusEnum = "RETRYING"
	statusFailure  JobStatusEnum = "FAILURE"
	statusSuccess  JobStatusEnum = "SUCCESS"
	statusQueueing JobStatusEnum = "QUEUEING"
	statusStopped  JobStatusEnum = "STOPPED"

	defaultInterval = 500 * time.Millisecond
)

var (
	registeredTask map[string]struct {
		handler     types.WorkerHandler
		workerIndex int
		moduleName  string
	}

	workers                []reflect.SelectCase
	runningWorkerIndexTask map[int]*Task

	queue      QueueStorage
	persistent Persistent

	refreshWorkerNotif, shutdown, closeAllSubscribers, semaphoreAddJob chan struct{}
	semaphore                                                          []chan struct{}
	mutex                                                              sync.Mutex
	tasks                                                              []string

	clientTaskSubscribers        map[string]*clientTaskDashboardSubscriber
	clientTaskJobListSubscribers map[string]*clientTaskJobListSubscriber
	clientJobDetailSubscribers   map[string]*clientJobDetailSubscriber

	errClientLimitExceeded = errors.New("client limit exceeded, please try again later")

	defaultOption option
)

func makeAllGlobalVars(service factory.ServiceFactory, opts ...OptionFunc) {

	// set default value
	defaultOption.tracingDashboard = "http://127.0.0.1:16686"
	defaultOption.maxClientSubscriber = 5
	defaultOption.autoRemoveClientInterval = 30 * time.Minute
	defaultOption.dashboardPort = 8080
	defaultOption.debugMode = true
	defaultOption.locker = &candiutils.NoopLocker{}
	defaultOption.dashboardBanner = `    _________    _   ______  ____
   / ____/   |  / | / / __ \/  _/
  / /   / /| | /  |/ / / / // /  
 / /___/ ___ |/ /|  / /_/ // /   
 \____/_/  |_/_/ |_/_____/___/   `

	//  override option value
	for _, opt := range opts {
		opt(&defaultOption)
	}

	// set default persistent & queue if not defined
	if defaultOption.persistent == nil {
		if service.GetDependency().GetMongoDatabase() != nil {
			defaultOption.persistent = NewMongoPersistent(service.GetDependency().GetMongoDatabase().WriteDB())
		} else if service.GetDependency().GetSQLDatabase() != nil {
			defaultOption.persistent = NewSQLPersistent(service.GetDependency().GetSQLDatabase().WriteDB())
		} else {
			defaultOption.persistent = NewNoopPersistent()
		}
	}

	if defaultOption.queue == nil {
		if service.GetDependency().GetRedisPool() != nil {
			defaultOption.queue = NewRedisQueue(service.GetDependency().GetRedisPool().WritePool())
		} else {
			defaultOption.queue = NewInMemQueue()
		}
	}

	queue = defaultOption.queue
	persistent = defaultOption.persistent

	refreshWorkerNotif, shutdown, closeAllSubscribers = make(chan struct{}), make(chan struct{}, 1), make(chan struct{})
	semaphoreAddJob = make(chan struct{}, env.BaseEnv().MaxGoroutines)
	clientTaskSubscribers = make(map[string]*clientTaskDashboardSubscriber, defaultOption.maxClientSubscriber)
	clientTaskJobListSubscribers = make(map[string]*clientTaskJobListSubscriber, defaultOption.maxClientSubscriber)
	clientJobDetailSubscribers = make(map[string]*clientJobDetailSubscriber, defaultOption.maxClientSubscriber)

	registeredTask = make(map[string]struct {
		handler     types.WorkerHandler
		workerIndex int
		moduleName  string
	})
	runningWorkerIndexTask = make(map[int]*Task)

	// add refresh worker channel to first index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(refreshWorkerNotif),
	})
}
