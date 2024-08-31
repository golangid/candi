package taskqueueworker

import (
	"context"
	"time"

	cronexpr "github.com/golangid/candi/candiutils/cronparser"
	"github.com/golangid/candi/codebase/factory/types"
)

type (
	// Task model
	Task struct {
		isInternalTask   bool
		internalTaskName string

		handler        types.WorkerHandler
		cancel         context.CancelFunc
		taskName       string
		moduleName     string
		workerIndex    int
		activeInterval *time.Ticker
		schedule       cronexpr.Schedule
	}

	// JobStatusEnum enum status
	JobStatusEnum string

	jobResult struct {
		err                         error
		result, traceID, stackTrace string
	}
)

// String method
func (j JobStatusEnum) String() string {
	return string(j)
}

const (
	defaultInterval = 500 * time.Millisecond

	// StatusRetrying const
	StatusRetrying JobStatusEnum = "RETRYING"
	// StatusFailure const
	StatusFailure JobStatusEnum = "FAILURE"
	// StatusSuccess const
	StatusSuccess JobStatusEnum = "SUCCESS"
	// StatusQueueing const
	StatusQueueing JobStatusEnum = "QUEUEING"
	// StatusStopped const
	StatusStopped JobStatusEnum = "STOPPED"
	// StatusHold const
	StatusHold JobStatusEnum = "HOLD"

	// HeaderRetries const
	HeaderRetries = "retries"
	// HeaderMaxRetries const
	HeaderMaxRetries = "max_retry"
	// HeaderInterval const
	HeaderInterval = "interval"
	// HeaderCurrentProgress const
	HeaderCurrentProgress = "current_progress"
	// HeaderMaxProgress const
	HeaderMaxProgress = "max_progress"

	// TaskOptionDeleteJobAfterSuccess const
	TaskOptionDeleteJobAfterSuccess = "delAfterSuccess"
)
