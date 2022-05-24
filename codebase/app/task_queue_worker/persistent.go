package taskqueueworker

import (
	"context"
)

type (
	// Persistent abstraction
	Persistent interface {
		SetSummary(Summary)
		Summary() Summary
		FindAllJob(ctx context.Context, filter *Filter) (jobs []Job)
		FindJobByID(ctx context.Context, id string, excludeFields ...string) (job Job, err error)
		CountAllJob(ctx context.Context, filter *Filter) int
		AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary)
		SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory)
		UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error)
		CleanJob(ctx context.Context, filter *Filter) (affectedRow int64)
		DeleteJob(ctx context.Context, id string) (job Job, err error)
	}

	noopPersistent struct {
		summary *inMemSummary
	}
)

// NewNoopPersistent constructor
func NewNoopPersistent() Persistent {

	values := make(map[string]*TaskSummary, len(tasks))
	for _, task := range tasks {
		values[task] = new(TaskSummary)
	}

	return &noopPersistent{
		summary: &inMemSummary{values: values},
	}
}

func (n *noopPersistent) SetSummary(s Summary) {
}
func (n *noopPersistent) Summary() Summary {
	return n.summary
}
func (n *noopPersistent) FindAllJob(ctx context.Context, filter *Filter) (jobs []Job) {
	return
}
func (n *noopPersistent) FindJobByID(ctx context.Context, id string, excludeFields ...string) (job Job, err error) {
	return
}
func (n *noopPersistent) CountAllJob(ctx context.Context, filter *Filter) (count int) {
	return
}
func (n *noopPersistent) AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary) {
	return
}
func (n *noopPersistent) SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) {
	return
}
func (n *noopPersistent) UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error) {
	return
}
func (n *noopPersistent) CleanJob(ctx context.Context, filter *Filter) (affectedRow int64) {
	return
}
func (n *noopPersistent) DeleteJob(ctx context.Context, id string) (job Job, err error) {
	return
}
