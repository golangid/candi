package taskqueueworker

import (
	"context"
	"sync"

	"github.com/golangid/candi/candishared"
)

// inMemQueue queue
type inMemQueue struct {
	mu    sync.Mutex
	queue map[string]*candishared.Queue[string]
}

// NewInMemQueue init inmem queue
func NewInMemQueue() QueueStorage {
	q := &inMemQueue{queue: make(map[string]*candishared.Queue[string])}
	return q
}

func (i *inMemQueue) PushJob(ctx context.Context, job *Job) (n int64) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.queue[job.TaskName] == nil {
		i.queue[job.TaskName] = candishared.NewQueue[string]()
	}
	i.queue[job.TaskName].Push(job.ID)
	return int64(i.queue[job.TaskName].Len())
}
func (i *inMemQueue) PopJob(ctx context.Context, taskName string) string {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.queue[taskName] == nil {
		i.queue[taskName] = candishared.NewQueue[string]()
	}
	el, _ := i.queue[taskName].Pop()
	return el
}
func (i *inMemQueue) NextJob(ctx context.Context, taskName string) string {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.queue[taskName] == nil {
		i.queue[taskName] = candishared.NewQueue[string]()
	}
	el, _ := i.queue[taskName].Peek()
	return el
}
func (i *inMemQueue) Clear(ctx context.Context, taskName string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.queue[taskName] = candishared.NewQueue[string]()
}
func (i *inMemQueue) Ping() error {
	return nil
}
func (i *inMemQueue) Type() string {
	return "In Memory Queue"
}
