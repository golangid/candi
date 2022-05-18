package taskqueueworker

import (
	"context"
	"strings"
)

// Persistent abstraction
type Persistent interface {
	FindAllJob(ctx context.Context, filter *Filter) (jobs []Job)
	FindJobByID(ctx context.Context, id string, excludeFields ...string) (job *Job, err error)
	CountAllJob(ctx context.Context, filter *Filter) int
	AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary)
	FindAllSummary(ctx context.Context, filter *Filter) (result []TaskSummary)
	UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{})
	IncrementSummary(ctx context.Context, taskName string, incr map[string]interface{})
	SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory)
	UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error)
	CleanJob(ctx context.Context, filter *Filter) (affectedRow int64)
	DeleteJob(ctx context.Context, id string) (job Job, err error)
}

type (
	// TaskSummary model
	TaskSummary struct {
		ID       string `bson:"_id"`
		TaskName string `bson:"task_name"`
		Success  int    `bson:"success"`
		Queueing int    `bson:"queueing"`
		Retrying int    `bson:"retrying"`
		Failure  int    `bson:"failure"`
		Stopped  int    `bson:"stopped"`
	}
)

// CountTotalJob method
func (s *TaskSummary) CountTotalJob() int {
	return s.Success + s.Queueing + s.Retrying + s.Failure + s.Stopped
}

// ToSummaryDetail method
func (s *TaskSummary) ToSummaryDetail() (detail SummaryDetail) {
	detail.Failure = s.Failure
	detail.Retrying = s.Retrying
	detail.Success = s.Success
	detail.Queueing = s.Queueing
	detail.Stopped = s.Stopped
	return
}

// ToMapResult method
func (s *TaskSummary) ToMapResult() map[string]int {
	return map[string]int{
		strings.ToUpper(string(statusFailure)):  s.Failure,
		strings.ToUpper(string(statusRetrying)): s.Retrying,
		strings.ToUpper(string(statusSuccess)):  s.Success,
		strings.ToUpper(string(statusQueueing)): s.Queueing,
		strings.ToUpper(string(statusStopped)):  s.Stopped,
	}
}

// SetValue method
func (s *TaskSummary) SetValue(source map[string]int) {
	s.Failure = source[strings.ToUpper(string(statusFailure))]
	s.Retrying = source[strings.ToUpper(string(statusRetrying))]
	s.Success = source[strings.ToUpper(string(statusSuccess))]
	s.Queueing = source[strings.ToUpper(string(statusQueueing))]
	s.Stopped = source[strings.ToUpper(string(statusStopped))]
}
