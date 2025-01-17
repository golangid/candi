package cache

import (
	"context"
	"fmt"
	"strings"
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
	defer func() { trace.Log("result", data); trace.Finish(tracer.FinishWithError(err)) }()

	trace.SetTag("db.statement", "GET")
	trace.SetTag("db.key", key)

	cl := r.read.Get()
	defer cl.Close()

	return redis.Bytes(cl.Do("GET", key))
}

// GetKeys method
func (r *RedisCache) GetKeys(ctx context.Context, pattern string) (data []string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:get_keys")
	defer func() { trace.Log("result", data); trace.Finish(tracer.FinishWithError(err)) }()

	trace.SetTag("db.statement", "KEYS")
	trace.SetTag("db.key", pattern)

	cl := r.read.Get()
	defer cl.Close()

	return redis.Strings(cl.Do("KEYS", fmt.Sprintf("%s*", pattern)))
}

// GetTTL method
func (r *RedisCache) GetTTL(ctx context.Context, key string) (dur time.Duration, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:get_ttl")
	defer func() { trace.Log("result", dur.String()); trace.Finish(tracer.FinishWithError(err)) }()

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
func (r *RedisCache) Set(ctx context.Context, key string, value any, expire time.Duration) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:set")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	trace.SetTag("db.statement", "SET")
	trace.SetTag("db.key", key)
	trace.SetTag("db.expired", expire.String())
	trace.Log("value", value)

	cl := r.write.Get()
	defer cl.Close()

	if _, err = cl.Do("SET", key, value); err != nil {
		return
	}

	if expire >= 0 {
		_, err = cl.Do("EXPIRE", key, int(expire.Seconds()))
	}
	return
}

// Exists method
func (r RedisCache) Exists(ctx context.Context, key string) (exist bool, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:exists")
	defer func() { trace.Log("result", exist); trace.Finish(tracer.FinishWithError(err)) }()

	trace.SetTag("db.statement", "EXISTS")
	trace.SetTag("db.key", key)

	cl := r.read.Get()
	defer cl.Close()

	return redis.Bool(cl.Do("EXISTS", key))
}

// Delete method with pattern
func (r *RedisCache) Delete(ctx context.Context, key string) (err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:delete")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	trace.SetTag("db.statement", "DEL")
	trace.SetTag("db.key", key)

	cl := r.write.Get()
	defer cl.Close()

	var keys []string
	if strings.HasSuffix(key, "*") {
		keys, _ = redis.Strings(cl.Do("KEYS", key))
	}

	if len(keys) == 0 {
		keys = []string{key}
	}

	for _, k := range keys {
		if _, err = cl.Do("DEL", k); err != nil {
			return err
		}
	}
	return nil
}

// DoCommand method to execute any Redis command
func (r *RedisCache) DoCommand(ctx context.Context, isWrite bool, command string, args ...any) (reply any, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "redis:do_command")
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	trace.SetTag("db.command", command)
	trace.Log("args", args)

	// Select the appropriate connection pool (read or write)
	cl := r.read.Get()
	if isWrite {
		cl = r.write.Get()
	}
	defer cl.Close()

	// Execute the Redis command
	return cl.Do(command, args...)
}
