package candiutils

import (
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
		pool *redis.Pool
	}

	// NoopLocker empty locker
	NoopLocker struct{}
)

// NewRedisLocker constructor
func NewRedisLocker(pool *redis.Pool) Locker {
	return &redisLocker{pool: pool}
}

func (r *redisLocker) IsLocked(key string) bool {
	conn := r.pool.Get()
	incr, _ := redis.Int64(conn.Do("INCR", key))
	conn.Close()

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
	keys, _ := redis.Strings(conn.Do("KEYS", key))
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
