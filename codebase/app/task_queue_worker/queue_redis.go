package taskqueueworker

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
)

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
