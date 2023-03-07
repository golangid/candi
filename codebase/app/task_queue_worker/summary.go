package taskqueueworker

import (
	"context"
	"strings"
	"sync"
)

type (
	// Summary abstraction
	Summary interface {
		FindAllSummary(ctx context.Context, filter *Filter) (result []TaskSummary)
		FindDetailSummary(ctx context.Context, taskName string) (result TaskSummary)
		UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{})
		IncrementSummary(ctx context.Context, taskName string, incr map[string]int64)
		DeleteAllSummary(ctx context.Context, filter *Filter)
	}
)

// summary store & read from in memory
type inMemSummary struct {
	mu     sync.Mutex
	values map[string]*TaskSummary
}

// NewInMemSummary constructor, store & read summary from in memory
func NewInMemSummary() Summary {

	values := make(map[string]*TaskSummary)
	return &inMemSummary{
		values: values,
	}
}

func (i *inMemSummary) FindAllSummary(ctx context.Context, filter *Filter) (result []TaskSummary) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for taskName, summary := range i.values {
		if filter.TaskName != "" && taskName != filter.TaskName {
			continue
		}
		summary.ID = taskName
		summary.TaskName = taskName
		result = append(result, *summary)
	}

	if len(filter.Statuses) > 0 {
		for i, res := range result {
			mapRes := res.ToMapResult()
			newCount := map[string]int{}
			for _, status := range filter.Statuses {
				newCount[strings.ToUpper(status)] = mapRes[strings.ToUpper(status)]
			}
			res.SetValue(newCount)
			result[i] = res
		}
	}
	return
}
func (i *inMemSummary) FindDetailSummary(ctx context.Context, taskName string) (result TaskSummary) {
	i.mu.Lock()
	defer i.mu.Unlock()

	summary := i.values[taskName]
	if summary == nil {
		summary = new(TaskSummary)
	}
	return *summary
}
func (i *inMemSummary) UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{}) {
	i.mu.Lock()
	defer i.mu.Unlock()

	summary := i.values[taskName]
	if summary == nil {
		summary = new(TaskSummary)
	}
	for k, v := range updated {
		var count int
		switch c := v.(type) {
		case int:
			count = c
		case int64:
			count = int(c)
		case bool:
			summary.IsLoading = c
		}
		switch strings.ToUpper(k) {
		case string(StatusFailure):
			summary.Failure = count
		case string(StatusRetrying):
			summary.Retrying = count
		case string(StatusSuccess):
			summary.Success = count
		case string(StatusQueueing):
			summary.Queueing = count
		case string(StatusStopped):
			summary.Stopped = count
		}
	}
	i.values[taskName] = summary
	return
}
func (i *inMemSummary) IncrementSummary(ctx context.Context, taskName string, incr map[string]int64) {
	i.mu.Lock()
	defer i.mu.Unlock()

	summary := i.values[taskName]
	if summary == nil {
		summary = new(TaskSummary)
	}
	for k, v := range incr {
		switch strings.ToUpper(k) {
		case string(StatusFailure):
			summary.Failure += int(v)
		case string(StatusRetrying):
			summary.Retrying += int(v)
		case string(StatusSuccess):
			summary.Success += int(v)
		case string(StatusQueueing):
			summary.Queueing += int(v)
		case string(StatusStopped):
			summary.Stopped += int(v)
		}
	}
	i.values[taskName] = summary
	return
}
func (i *inMemSummary) DeleteAllSummary(ctx context.Context, filter *Filter) {}
