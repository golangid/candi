package candiutils

import (
	"context"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Lock implementation of interfaces.Locker, lock concurrent process either in one runtime or multiple runtimes

type (
	// RedisLocker lock using redis
	RedisLocker struct {
		pool *redis.Pool
	}

	// NoopLocker empty locker
	NoopLocker struct{}
)

// NewRedisLocker constructor
func NewRedisLocker(pool *redis.Pool) *RedisLocker {
	return &RedisLocker{pool: pool}
}

func (r *RedisLocker) IsLocked(key string) bool {
	conn := r.pool.Get()
	incr, _ := redis.Int64(conn.Do("INCR", key))
	conn.Close()

	return incr > 1
}

func (r *RedisLocker) HasBeenLocked(key string) bool {
	conn := r.pool.Get()
	incr, _ := redis.Int64(conn.Do("GET", key))
	conn.Close()

	return incr > 0
}

// Unlock method
func (r *RedisLocker) Unlock(key string) {
	conn := r.pool.Get()
	conn.Do("DEL", key)
	conn.Close()
}

// Reset method
func (r *RedisLocker) Reset(key string) {
	conn := r.pool.Get()
	keys, _ := redis.Strings(conn.Do("KEYS", key))
	for _, k := range keys {
		conn.Do("DEL", k)
	}
	conn.Close()
	return
}

// Disconnect close and reset
func (r *RedisLocker) Disconnect(ctx context.Context) error {
	conn := r.pool.Get()
	conn.Do("DEL", "LOCKFOR:*")
	conn.Close()
	return nil
}

// Lock method
func (r *RedisLocker) Lock(key string, timeout time.Duration) (unlockFunc func(), err error) {
	if timeout <= 0 {
		return func() {}, errors.New("Timeout must be positive")
	}
	if key == "" {
		return func() {}, errors.New("Key cannot empty")
	}

	lockKey := "LOCKFOR:" + key
	unlockFunc = func() { r.Unlock(lockKey) }
	if !r.IsLocked(lockKey) {
		return unlockFunc, nil
	}

	conn := r.pool.Get()
	conn.Do("CONFIG", "SET", "notify-keyspace-events", "KEA")

	eventChannel := "__key*__:del"
	psc := &redis.PubSubConn{Conn: conn}
	psc.PSubscribe(eventChannel)
	defer func() { psc.Unsubscribe(); conn.Close() }()

	wait := make(chan error)
	go func() {
		for {
			switch msg := psc.Receive().(type) {
			case redis.Message:
				if msg.Pattern == eventChannel && lockKey == string(msg.Data) && !r.IsLocked(lockKey) {
					wait <- nil
					return
				}

			case error:
				wait <- msg
				return
			}
		}
	}()

	select {
	case err := <-wait:
		return unlockFunc, err

	case <-time.After(timeout):
		r.Unlock(lockKey)
		return unlockFunc, errors.New("Timeout when waiting unlock another process")
	}
}

// NoopLocker

// IsLocked method
func (NoopLocker) IsLocked(string) bool { return false }

// HasBeenLocked method
func (NoopLocker) HasBeenLocked(string) bool { return false }

// Unlock method
func (NoopLocker) Unlock(string) {}

// Reset method
func (NoopLocker) Reset(string) {}

// Lock method
func (NoopLocker) Lock(string, time.Duration) (func(), error) { return func() {}, nil }

func (NoopLocker) Disconnect(context.Context) error { return nil }
