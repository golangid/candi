package candiutils

import (
	"github.com/gomodule/redigo/redis"
)

type (
	// Locker abstraction, lock concurrent processs
	Locker interface {
		IsLocked(key string) bool
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

func (r *redisLocker) IsLocked(key string) bool {
	incr, _ := redis.Int64(r.conn.Do("INCR", key))
	defer func() {
		if incr <= 1 {
			r.conn.Do("DEL", key)
		}
	}()

	return incr > 1
}

// IsLocked method
func (NoopLocker) IsLocked(key string) bool {
	return false
}
