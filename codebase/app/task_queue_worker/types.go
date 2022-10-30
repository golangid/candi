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
)

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

const (
	// TaskOptionDeleteJobAfterSuccess const
	TaskOptionDeleteJobAfterSuccess = "delAfterSuccess"
)
