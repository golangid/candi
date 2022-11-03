package taskqueueworker

import (
	"context"
	"errors"

	"github.com/golangid/candi/tracer"
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

func (r *redisQueue) PushJob(ctx context.Context, job *Job) (n int64) {
	tracer.Log(ctx, "redis.queue:push_job", job.ID)

	conn := r.pool.Get()
	defer conn.Close()

	n, _ = redis.Int64(conn.Do("RPUSH", job.TaskName, job.ID))
	return
}
func (r *redisQueue) PopJob(ctx context.Context, taskName string) string {
	conn := r.pool.Get()
	defer conn.Close()

	id, _ := redis.String(conn.Do("LPOP", taskName))
	return id
}
func (r *redisQueue) NextJob(ctx context.Context, taskName string) string {
	tracer.Log(ctx, "redis.queue:next_job", taskName)

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
func (r *redisQueue) Clear(ctx context.Context, taskName string) {
	conn := r.pool.Get()
	defer conn.Close()

	conn.Do("DEL", taskName)
}
func (r *redisQueue) Ping() error {

	ping := r.pool.Get()
	defer ping.Close()

	_, err := ping.Do("PING")
	if err != nil {
		return errors.New("redis ping: " + err.Error())
	}
	return nil
}
func (r *redisQueue) Type() string {
	return "Redis Queue"
}
