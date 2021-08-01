package taskqueueworker

import (
	"encoding/json"
	"sync"

	"github.com/gomodule/redigo/redis"
	"pkg.agungdp.dev/candi/candishared"
)

// QueueStorage abstraction for queue storage backend
type QueueStorage interface {
	PushJob(job *Job)
	PopJob(taskName string) string
	NextJob(taskName string) string
	Clear(taskName string)
}

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

// redisQueue queue
type redisQueue struct {
	pool *redis.Pool
}

// NewRedisQueue init inmem queue
func NewRedisQueue(redisPool *redis.Pool) QueueStorage {
	if redisPool == nil {
		panic("Task queue backend require redis")
	}
	return &redisQueue{pool: redisPool}
}

func (r *redisQueue) GetAllJobs(taskName string) (jobs []*Job) {
	conn := r.pool.Get()
	defer conn.Close()

	str, _ := conn.Do("LRANGE", taskName, 0, -1)
	results, _ := str.([]interface{})
	for _, result := range results {
		b, _ := result.([]byte)
		var job Job
		json.Unmarshal(b, &job)
		jobs = append(jobs, &job)
	}
	return
}
func (r *redisQueue) PushJob(job *Job) {
	conn := r.pool.Get()
	defer conn.Close()

	conn.Do("RPUSH", job.TaskName, job.ID)
}
func (r *redisQueue) PopJob(taskName string) string {
	conn := r.pool.Get()
	defer conn.Close()

	id, _ := redis.String(conn.Do("LPOP", taskName))
	return id
}
func (r *redisQueue) NextJob(taskName string) string {
	conn := r.pool.Get()
	defer conn.Close()

	id, err := redis.String(conn.Do("LINDEX", taskName, 0))
	if err != nil {
		return ""
	}

	if len(id) == 0 {
		return ""
	}
	return id
}
func (r *redisQueue) Clear(taskName string) {
	conn := r.pool.Get()
	defer conn.Close()

	conn.Do("DEL", taskName)
}
