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
	}

	// TaskSummary model
	TaskSummary struct {
		ID        string `bson:"_id"`
		TaskName  string `bson:"task_name"`
		Success   int    `bson:"success"`
		Queueing  int    `bson:"queueing"`
		Retrying  int    `bson:"retrying"`
		Failure   int    `bson:"failure"`
		Stopped   int    `bson:"stopped"`
		IsLoading bool   `bson:"is_loading"`
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

// summary store & read from in memory
type inMemSummary struct {
	mu     sync.Mutex
	values map[string]*TaskSummary
}

// NewInMemSummary constructor, store & read summary from in memory
func NewInMemSummary() Summary {

	values := make(map[string]*TaskSummary, len(tasks))
	for _, task := range tasks {
		values[task] = new(TaskSummary)
	}

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
		case string(statusFailure):
			summary.Failure = count
		case string(statusRetrying):
			summary.Retrying = count
		case string(statusSuccess):
			summary.Success = count
		case string(statusQueueing):
			summary.Queueing = count
		case string(statusStopped):
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
		case string(statusFailure):
			summary.Failure += int(v)
		case string(statusRetrying):
			summary.Retrying += int(v)
		case string(statusSuccess):
			summary.Success += int(v)
		case string(statusQueueing):
			summary.Queueing += int(v)
		case string(statusStopped):
			summary.Stopped += int(v)
		}
	}
	i.values[taskName] = summary
	return
}
