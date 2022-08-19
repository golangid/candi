package taskqueueworker

import (
	"context"
)

type (
	// Persistent abstraction
	Persistent interface {
		Ping(ctx context.Context) error
		SetSummary(Summary)
		Summary() Summary
		FindAllJob(ctx context.Context, filter *Filter) (jobs []Job)
		FindJobByID(ctx context.Context, id string, filterHistory *Filter) (job Job, err error)
		CountAllJob(ctx context.Context, filter *Filter) int
		AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary)
		SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) (err error)
		UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error)
		CleanJob(ctx context.Context, filter *Filter) (affectedRow int64)
		DeleteJob(ctx context.Context, id string) (job Job, err error)
		Type() string
		GetAllConfiguration(ctx context.Context) (cfg []Configuration, err error)
		GetConfiguration(key string) (Configuration, error)
		SetConfiguration(cfg *Configuration) (err error)
	}

	noopPersistent struct {
		summary *inMemSummary
	}
)

// NewNoopPersistent constructor
func NewNoopPersistent() Persistent {

	values := make(map[string]*TaskSummary)

	return &noopPersistent{
		summary: &inMemSummary{values: values},
	}
}

func (n *noopPersistent) Ping(ctx context.Context) error {
	return nil
}
func (n *noopPersistent) SetSummary(s Summary) {
}
func (n *noopPersistent) Summary() Summary {
	return n.summary
}
func (n *noopPersistent) FindAllJob(ctx context.Context, filter *Filter) (jobs []Job) {
	return
}
func (n *noopPersistent) FindJobByID(ctx context.Context, id string, filter *Filter) (job Job, err error) {
	return
}
func (n *noopPersistent) CountAllJob(ctx context.Context, filter *Filter) (count int) {
	return
}
func (n *noopPersistent) AggregateAllTaskJob(ctx context.Context, filter *Filter) (result []TaskSummary) {
	return
}
func (n *noopPersistent) SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) (err error) {
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
func (n *noopPersistent) Type() string {
	return "In Memory Persistent"
}
func (n *noopPersistent) GetAllConfiguration(ctx context.Context) (cfg []Configuration, err error) {
	return
}
func (n *noopPersistent) GetConfiguration(key string) (cfg Configuration, err error) { return }
func (n *noopPersistent) SetConfiguration(cfg *Configuration) (err error)            { return }
