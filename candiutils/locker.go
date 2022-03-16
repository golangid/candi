package candiutils

import (
	"time"

	"github.com/golangid/candi/config/env"
	"github.com/gomodule/redigo/redis"
)

type (
	// Locker abstraction, lock concurrent processs
	Locker interface {
		IsLocked(key string) bool
		Unlock(key string)
		Reset(key string)
	}

	redisLocker struct {
		pool          *redis.Pool
		lockerTimeout time.Duration
	}

	// NoopLocker empty locker
	NoopLocker struct{}
)

// NewRedisLocker constructor
func NewRedisLocker(pool *redis.Pool) Locker {
	return &redisLocker{
		pool:          pool,
		lockerTimeout: env.BaseEnv().LockerTimeout,
	}
}

func (r *redisLocker) IsLocked(key string) bool {
	conn := r.pool.Get()
	defer func() { conn.Close() }()

	// set atomic transaction
	conn.Send("MULTI")
	conn.Send("INCR", key)
	conn.Send("EXPIRE", key, r.lockerTimeout.Seconds())

	vals, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return false
	}

	if len(vals) <= 0 {
		return false
	}

	incr, ok := vals[0].(int64)
	if !ok {
		return false
	}

	return incr > 1
}

// Unlock method
func (r *redisLocker) Unlock(key string) {
	conn := r.pool.Get()
	conn.Do("DEL", key)
	conn.Close()
}

// Reset method
func (r *redisLocker) Reset(key string) {
	conn := r.pool.Get()
	keys, err := redis.Strings(conn.Do("KEYS", key))
	if err != nil {
		return
	}

	for _, k := range keys {
		conn.Do("DEL", k)
	}
	conn.Close()
	return
}

// IsLocked method
func (NoopLocker) IsLocked(key string) bool {
	return false
}

// Unlock method
func (NoopLocker) Unlock(key string) {
}

// Reset method
func (NoopLocker) Reset(key string) {
}
