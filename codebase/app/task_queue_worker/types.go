package taskqueueworker

import (
	"context"
	"time"
)

type (
	// Task model
	Task struct {
		isInternalTask   bool
		internalTaskName string

		cancel         context.CancelFunc
		taskName       string
		workerIndex    int
		activeInterval *time.Ticker
		nextInterval   *time.Duration
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
