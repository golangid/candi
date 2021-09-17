package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/golangid/candi/tracer"
	"github.com/gomodule/redigo/redis"
)

// RedisCache redis implement interfaces.Cache
type RedisCache struct {
	read, write *redis.Pool
}

// NewRedisCache constructor
func NewRedisCache(read, write *redis.Pool) *RedisCache {
	return &RedisCache{read: read, write: write}
}

// Get method
func (r *RedisCache) Get(ctx context.Context, key string) (data []byte, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:get")
	defer func() { trace.SetError(err); trace.Log("result", data); trace.Finish() }()

	trace.SetTag("db.statement", "GET")
	trace.SetTag("db.key", key)

	cl := r.read.Get()
	defer cl.Close()

	return redis.Bytes(cl.Do("GET", key))
}

// GetKeys method
func (r *RedisCache) GetKeys(ctx context.Context, pattern string) (data []string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:get_keys")
	defer func() { trace.SetError(err); trace.Log("result", data); trace.Finish() }()

	trace.SetTag("db.statement", "KEYS")
	trace.SetTag("db.key", pattern)

	cl := r.read.Get()
	defer cl.Close()

	return redis.Strings(cl.Do("KEYS", fmt.Sprintf("%s*", pattern)))
}

// GetTTL method
func (r *RedisCache) GetTTL(ctx context.Context, key string) (dur time.Duration, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:get_ttl")
	defer func() { trace.SetError(err); trace.Log("result", dur.String()); trace.Finish() }()

	trace.SetTag("db.statement", "GetTTL")
	trace.SetTag("db.key", key)

	cl := r.read.Get()
	defer cl.Close()

	reply, err := cl.Do("TTL", key)
	if err != nil {
		return dur, err
	}

	sec, _ := reply.(int64)
	return time.Duration(sec) * time.Second, nil
}

// Set method
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expire time.Duration) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:set")
	defer func() { trace.SetError(err); trace.Finish() }()

	trace.SetTag("db.statement", "SET")
	trace.SetTag("db.key", key)
	trace.SetTag("db.expired", expire.String())
	trace.Log("value", value)

	cl := r.write.Get()
	defer cl.Close()

	if _, err = cl.Do("SET", key, value); err != nil {
		return
	}
	if _, err = cl.Do("EXPIRE", key, int(expire.Seconds())); err != nil {
		return
	}
	return nil
}

// Exists method
func (r RedisCache) Exists(ctx context.Context, key string) (exist bool, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:exists")
	defer func() { trace.SetError(err); trace.Log("result", exist); trace.Finish() }()

	trace.SetTag("db.statement", "EXISTS")
	trace.SetTag("db.key", key)

	cl := r.read.Get()
	defer cl.Close()

	return redis.Bool(cl.Do("EXISTS", key))
}

// Delete method
func (r *RedisCache) Delete(ctx context.Context, key string) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:delete")
	defer func() { trace.SetError(err); trace.Finish() }()

	trace.SetTag("db.statement", "DEL")
	trace.SetTag("db.key", key)

	cl := r.write.Get()
	defer cl.Close()

	_, err = cl.Do("DEL", key)
	return nil
}
