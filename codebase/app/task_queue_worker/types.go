package taskqueueworker

import (
	"errors"
	"net/url"
	"reflect"
	"sync"
	"time"

	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config/env"
)

type (
	// TaglineResolver resolver
	TaglineResolver struct {
		Tagline                   string
		TaskListClientSubscribers []string
		JobListClientSubscribers  []string
	}
	// TaskResolver resolver
	TaskResolver struct {
		Name      string
		TotalJobs int
		Detail    struct {
			GiveUp, Retrying, Success, Queueing, Stopped int
		}
	}

	// Meta resolver
	Meta struct {
		Page         int
		Limit        int
		TotalRecords int
		TotalPages   int
		Detail       struct {
			GiveUp, Retrying, Success, Queueing, Stopped int
		}
	}

	// JobListResolver resolver
	JobListResolver struct {
		Meta Meta
		Data []Job
	}

	// Filter type
	Filter struct {
		Page, Limit int
		TaskName    string
		Search      *string
		Status      []string
	}

	clientJobTaskSubscriber struct {
		c      chan JobListResolver
		filter Filter
	}

	jobStatusEnum string
)

const (
	statusRetrying jobStatusEnum = "RETRYING"
	statusFailure  jobStatusEnum = "FAILURE"
	statusSuccess  jobStatusEnum = "SUCCESS"
	statusQueueing jobStatusEnum = "QUEUEING"
	statusStopped  jobStatusEnum = "STOPPED"
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
	repo                                    *storage
	refreshWorkerNotif, shutdown, semaphore chan struct{}
	mutex                                   sync.Mutex
	tasks                                   []string
	tracerHost                              string

	clientTaskSubscribers    map[string]chan []TaskResolver
	clientJobTaskSubscribers map[string]clientJobTaskSubscriber

	errClientLimitExceeded = errors.New("client limit exceeded, please try again later")
)

func makeAllGlobalVars(service factory.ServiceFactory) {
	if service.GetDependency().GetRedisPool() == nil {
		panic("Task queue worker require redis for queue")
	}
	if service.GetDependency().GetMongoDatabase() == nil {
		panic("Task queue worker require mongo for dashboard management")
	}

	createMongoIndex(service.GetDependency().GetMongoDatabase().WriteDB())
	queue = NewRedisQueue(service.GetDependency().GetRedisPool().WritePool())
	repo = &storage{mongoRead: service.GetDependency().GetMongoDatabase().ReadDB(), mongoWrite: service.GetDependency().GetMongoDatabase().WriteDB()}
	refreshWorkerNotif, shutdown, semaphore = make(chan struct{}), make(chan struct{}, 1), make(chan struct{}, env.BaseEnv().MaxGoroutines)
	if urlTracerAgent, _ := url.Parse("//" + env.BaseEnv().JaegerTracingHost); urlTracerAgent != nil {
		tracerHost = urlTracerAgent.Hostname()
	}

	clientTaskSubscribers = make(map[string]chan []TaskResolver, env.BaseEnv().TaskQueueDashboardMaxClientSubscribers)
	clientJobTaskSubscribers = make(map[string]clientJobTaskSubscriber, env.BaseEnv().TaskQueueDashboardMaxClientSubscribers)

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
}
