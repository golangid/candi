package taskqueueworker

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"pkg.agungdwiprasetyo.com/candi/candishared"
)

// QueueStorage abstraction for queue storage backend
type QueueStorage interface {
	GetAllJobs(taskName string) []*Job
	PushJob(job *Job)
	PopJob(taskName string) Job
	NextJob(taskName string) *Job
	Clear(taskName string)
}

// inMemQueue queue
type inMemQueue struct {
	queue map[string]*candishared.Queue
}

// NewInMemQueue init inmem queue
func NewInMemQueue() QueueStorage {
	q := &inMemQueue{queue: make(map[string]*candishared.Queue)}
	return q
}

func (i *inMemQueue) GetAllJobs(taskName string) (jobs []*Job) {
	return nil
}
func (i *inMemQueue) PushJob(job *Job) {
	defer func() { recover() }()
	queue := i.queue[job.TaskName]
	if queue == nil {
		queue = candishared.NewQueue()
	}
	queue.Push(job)
}
func (i *inMemQueue) PopJob(taskName string) Job {
	defer func() { recover() }()
	return *i.queue[taskName].Pop().(*Job)
}
func (i *inMemQueue) NextJob(taskName string) *Job {
	defer func() { recover() }()
	return i.queue[taskName].Peek().(*Job)
}
func (i *inMemQueue) Clear(taskName string) {
	defer func() { recover() }()
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
func (r *redisQueue) PopJob(taskName string) Job {
	conn := r.pool.Get()
	defer conn.Close()

	var job Job
	job.ID, _ = redis.String(conn.Do("LPOP", taskName))
	return job
}
func (r *redisQueue) NextJob(taskName string) *Job {
	conn := r.pool.Get()
	defer conn.Close()

	b, err := redis.String(conn.Do("LINDEX", taskName, 0))
	if err != nil {
		return nil
	}

	if len(b) == 0 {
		return nil
	}

	var job Job
	job.ID = b
	return &job
}
func (r *redisQueue) Clear(taskName string) {
	conn := r.pool.Get()
	defer conn.Close()

	conn.Do("DEL", taskName)
}
