package taskqueueworker

import (
	"sync"

	"pkg.agungdp.dev/candi/candishared"
)

// inMemQueue queue
type inMemQueue struct {
	mu    sync.Mutex
	queue map[string]*candishared.Queue
}

// NewInMemQueue init inmem queue
func NewInMemQueue() QueueStorage {
	q := &inMemQueue{queue: make(map[string]*candishared.Queue)}
	return q
}

func (i *inMemQueue) PushJob(job *Job) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.queue[job.TaskName] == nil {
		i.queue[job.TaskName] = candishared.NewQueue()
	}
	i.queue[job.TaskName].Push(job.ID)
}
func (i *inMemQueue) PopJob(taskName string) string {
	el, err := i.queue[taskName].Pop()
	if err != nil {
		return ""
	}

	id, ok := el.(string)
	if !ok {
		return ""
	}
	return id
}
func (i *inMemQueue) NextJob(taskName string) string {

	el, err := i.queue[taskName].Peek()
	if err != nil {
		return ""
	}
	id, ok := el.(string)
	if !ok {
		return ""
	}
	return id
}
func (i *inMemQueue) Clear(taskName string) {

	i.queue[taskName] = nil
}
