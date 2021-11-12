package candiutils

import (
	"github.com/gomodule/redigo/redis"
)

type (
	// Locker abstraction, lock concurrent processs
	Locker interface {
		IsLocked(key string) (isLock bool, releaseLock func())
	}

	redisLocker struct {
		conn redis.Conn
	}

	// NoopLocker empty locker
	NoopLocker struct{}
)

// NewRedisLocker constructor
func NewRedisLocker(conn redis.Conn) Locker {
	return &redisLocker{conn: conn}
}

func (r *redisLocker) IsLocked(key string) (bool, func()) {
	incr, _ := redis.Int64(r.conn.Do("INCR", key))

	return incr > 1, func() { r.conn.Do("DEL", key) }
}

// IsLocked method
func (NoopLocker) IsLocked(key string) (bool, func()) {
	return false, func() {}
}
