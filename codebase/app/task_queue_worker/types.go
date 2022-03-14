package taskqueueworker

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory/types"
)

type (
	// Task model
	Task struct {
		cancel         context.CancelFunc
		taskName       string
		activeInterval *time.Ticker
	}

	// TaglineResolver resolver
	TaglineResolver struct {
		Banner                    string
		Tagline                   string
		Version                   string
		StartAt                   string
		BuildNumber               string
		TaskListClientSubscribers []string
		JobListClientSubscribers  []string
		MemoryStatistics          MemstatsResolver
	}
	// MemstatsResolver resolver
	MemstatsResolver struct {
		Alloc         string
		TotalAlloc    string
		NumGC         int
		NumGoroutines int
	}
	// MetaTaskResolver meta resolver
	MetaTaskResolver struct {
		Page                  int
		Limit                 int
		TotalRecords          int
		TotalPages            int
		IsCloseSession        bool
		TotalClientSubscriber int
	}
	// TaskResolver resolver
	TaskResolver struct {
		Name       string
		ModuleName string
		TotalJobs  int
		Detail     struct {
			Failure, Retrying, Success, Queueing, Stopped int
		}
	}
	// TaskListResolver resolver
	TaskListResolver struct {
		Meta MetaTaskResolver
		Data []TaskResolver
	}

	// MetaJobList resolver
	MetaJobList struct {
		Page           int
		Limit          int
		TotalRecords   int
		TotalPages     int
		IsCloseSession bool
		Detail         struct {
			Failure, Retrying, Success, Queueing, Stopped int
		}
	}

	// JobListResolver resolver
	JobListResolver struct {
		Meta MetaJobList
		Data []Job
	}

	// Filter type
	Filter struct {
		Page, Limit        int
		TaskName           string
		TaskNameList       []string
		Search, JobID      *string
		Status             []string
		ShowAll            bool
		ShowHistories      *bool
		StartDate, EndDate time.Time
	}

	// ClientSubscriber model
	ClientSubscriber struct {
		ClientID      string
		SubscribeList struct {
			TaskDashboard bool
			JobDetailID   string
			JobList       Filter
		}
	}

	clientTaskJobListSubscriber struct {
		c      chan JobListResolver
		filter Filter
	}
	clientJobDetailSubscriber struct {
		c     chan Job
		jobID string
	}

	// JobStatusEnum enum status
	JobStatusEnum string
)

const (
	statusRetrying JobStatusEnum = "RETRYING"
	statusFailure  JobStatusEnum = "FAILURE"
	statusSuccess  JobStatusEnum = "SUCCESS"
	statusQueueing JobStatusEnum = "QUEUEING"
	statusStopped  JobStatusEnum = "STOPPED"
)

var (
	registeredTask map[string]struct {
		handler     types.WorkerHandler
		workerIndex int
		moduleName  string
	}

	workers         []reflect.SelectCase
	workerIndexTask map[int]*Task

	queue                                             QueueStorage
	persistent                                        Persistent
	refreshWorkerNotif, shutdown, closeAllSubscribers chan struct{}
	semaphore                                         []chan struct{}
	mutex                                             sync.Mutex
	tasks                                             []string

	clientTaskSubscribers        map[string]chan TaskListResolver
	clientTaskJobListSubscribers map[string]clientTaskJobListSubscriber
	clientJobDetailSubscribers   map[string]clientJobDetailSubscriber

	errClientLimitExceeded = errors.New("client limit exceeded, please try again later")

	defaultOption option
)

func makeAllGlobalVars(q QueueStorage, perst Persistent, opts ...OptionFunc) {

	queue = q
	persistent = perst

	// set default value
	defaultOption.tracingDashboard = "http://127.0.0.1:16686"
	defaultOption.maxClientSubscriber = 5
	defaultOption.autoRemoveClientInterval = 30 * time.Minute
	defaultOption.dashboardPort = 8080
	defaultOption.debugMode = true
	defaultOption.locker = &candiutils.NoopLocker{}
	defaultOption.dashboardBanner = `
    _________    _   ______  ____
   / ____/   |  / | / / __ \/  _/
  / /   / /| | /  |/ / / / // /  
 / /___/ ___ |/ /|  / /_/ // /   
 \____/_/  |_/_/ |_/_____/___/   `

	//  override option value
	for _, opt := range opts {
		opt(&defaultOption)
	}

	refreshWorkerNotif, shutdown, closeAllSubscribers = make(chan struct{}), make(chan struct{}, 1), make(chan struct{})
	clientTaskSubscribers = make(map[string]chan TaskListResolver, defaultOption.maxClientSubscriber)
	clientTaskJobListSubscribers = make(map[string]clientTaskJobListSubscriber, defaultOption.maxClientSubscriber)
	clientJobDetailSubscribers = make(map[string]clientJobDetailSubscriber, defaultOption.maxClientSubscriber)

	registeredTask = make(map[string]struct {
		handler     types.WorkerHandler
		workerIndex int
		moduleName  string
	})
	workerIndexTask = make(map[int]*Task)

	// add refresh worker channel to first index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(refreshWorkerNotif),
	})
}
