package taskqueueworker

import (
	"encoding/json"

	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"github.com/gomodule/redigo/redis"
)

// QueueStorage abstraction for queue storage backend
type QueueStorage interface {
	GetAllJobs(taskID string) []*Job
	PushJob(job *Job)
	PopJob(taskID string) *Job
	NextJob(taskID string) *Job
}

// inMemQueue queue
type inMemQueue struct {
	queue map[string]*shared.Queue
}

// NewInMemQueue init inmem queue
func NewInMemQueue() QueueStorage {
	q := &inMemQueue{queue: make(map[string]*shared.Queue)}
	return q
}

func (i *inMemQueue) GetAllJobs(taskID string) (jobs []*Job) {
	return nil
}
func (i *inMemQueue) PushJob(job *Job) {
	defer func() { recover() }()
	queue := i.queue[job.TaskID]
	if queue == nil {
		queue = shared.NewQueue()
	}
	queue.Push(job)
}
func (i *inMemQueue) PopJob(taskID string) *Job {
	defer func() { recover() }()
	return i.queue[taskID].Pop().(*Job)
}
func (i *inMemQueue) NextJob(taskID string) *Job {
	defer func() { recover() }()
	return i.queue[taskID].Peek().(*Job)
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

func (r *redisQueue) GetAllJobs(taskID string) (jobs []*Job) {
	conn := r.pool.Get()
	defer conn.Close()

	str, _ := conn.Do("LRANGE", taskID, 0, -1)
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

	conn.Do("RPUSH", job.TaskID, helper.ToBytes(job))
}
func (r *redisQueue) PopJob(taskID string) *Job {
	conn := r.pool.Get()
	defer conn.Close()

	b, _ := redis.Bytes(conn.Do("LPOP", taskID))

	var job Job
	json.Unmarshal(b, &job)
	return &job
}
func (r *redisQueue) NextJob(taskID string) *Job {
	conn := r.pool.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("LINDEX", taskID, 0))
	if err != nil {
		return nil
	}

	if len(b) == 0 {
		return nil
	}

	var job Job
	json.Unmarshal(b, &job)
	return &job
}
