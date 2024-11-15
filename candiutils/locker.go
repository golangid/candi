package candiutils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Lock implementation of interfaces.Locker, lock concurrent process either in one runtime or multiple runtimes

type (
	// RedisLocker lock using redis
	RedisLocker struct {
		pool          *redis.Pool
		lockeroptions LockerOptions
	}

	// NoopLocker empty locker
	NoopLocker struct{}

	// Options for RedisLocker
	LockerOptions struct {
		Prefix string
		TTL    time.Duration
	}

	// Option function type for setting options
	LockerOption func(*LockerOptions)
)

// WithPrefix sets the prefix for keys
func WithPrefixLocker(prefix string) LockerOption {
	return func(o *LockerOptions) {
		o.Prefix = prefix
	}
}

// WithTTL sets the default TTL for keys
func WithTTLLocker(ttl time.Duration) LockerOption {
	return func(o *LockerOptions) {
		o.TTL = ttl
	}
}

// NewRedisLocker constructor
func NewRedisLocker(pool *redis.Pool, opts ...LockerOption) *RedisLocker {
	lockeroptions := LockerOptions{
		Prefix: "LOCKFOR",
		TTL:    0,
	}
	for _, opt := range opts {
		opt(&lockeroptions)
	}
	return &RedisLocker{pool: pool, lockeroptions: lockeroptions}
}

// GetPrefix returns the prefix used for keys
func (r *RedisLocker) GetPrefixLocker() string {
	return r.lockeroptions.Prefix + ":"
}

// GetTTLLocker returns the default TTL for keys
func (r *RedisLocker) GetTTLLocker() time.Duration {
	return r.lockeroptions.TTL
}

func (r *RedisLocker) IsLocked(key string) bool {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := fmt.Sprintf("%s:%s", r.lockeroptions.Prefix, key)
	incr, err := redis.Int64(conn.Do("INCR", lockKey))
	if err != nil {
		return false
	}

	return incr > 1
}

func (r *RedisLocker) IsLockedTTL(key string, TTL time.Duration) bool {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := fmt.Sprintf("%s:%s", r.lockeroptions.Prefix, key)
	incr, err := redis.Int64(conn.Do("INCR", lockKey))
	if err != nil {
		return false
	}

	var expireTime time.Duration
	if TTL > 0 {
		expireTime = TTL
	} else {
		expireTime = r.lockeroptions.TTL
	}

	if expireTime > 0 {
		conn.Do("EXPIRE", lockKey, int(expireTime.Seconds()))
	}

	return incr > 1
}

// IsLockedTTLWithLimit checks if the key has been incremented more than the specified limit
// within the given TTL. If the key is being created for the first time, it sets the TTL.
// Example usage: check if a key has been incremented more than 10 times within 1 minute.
func (r *RedisLocker) IsLockedTTLWithLimit(key string, limit int, TTL time.Duration) bool {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := fmt.Sprintf("%s:%s", r.lockeroptions.Prefix, key)
	incr, err := redis.Int64(conn.Do("INCR", lockKey))
	if err != nil {
		return false
	}

	var expireTime time.Duration
	if TTL > 0 {
		expireTime = TTL
	} else {
		expireTime = r.lockeroptions.TTL
	}

	if expireTime > 0 && incr == 1 {
		conn.Do("EXPIRE", lockKey, int(expireTime.Seconds()))
	}

	return incr > int64(limit)
}

func (r *RedisLocker) HasBeenLocked(key string) bool {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := fmt.Sprintf("%s:%s", r.lockeroptions.Prefix, key)
	incr, _ := redis.Int64(conn.Do("GET", lockKey))

	return incr > 0
}

// Unlock method
func (r *RedisLocker) Unlock(key string) {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := fmt.Sprintf("%s:%s", r.lockeroptions.Prefix, key)
	conn.Do("DEL", lockKey)
}

// Reset method
func (r *RedisLocker) Reset(key string) {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := fmt.Sprintf("%s:%s", r.lockeroptions.Prefix, key)
	keys, err := redis.Strings(conn.Do("KEYS", lockKey))
	if err != nil {
		fmt.Println("Error when reset locker: ", key, err)
		return
	}

	for _, k := range keys {
		_, err := conn.Do("DEL", k)
		if err != nil {
			fmt.Println("Error when reset locker: ", key, err)
		}
	}
}

// Disconnect close and reset
func (r *RedisLocker) Disconnect(ctx context.Context) error {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := fmt.Sprintf("%s:*", r.lockeroptions.Prefix)
	_, err := conn.Do("DEL", lockKey)
	if err != nil {
		return err
	}

	return nil
}

// Lock method
func (r *RedisLocker) Lock(key string, timeout time.Duration) (unlockFunc func(), err error) {
	if timeout <= 0 {
		return func() {}, errors.New("timeout must be positive")
	}
	if key == "" {
		return func() {}, errors.New("key cannot empty")
	}

	lockKey := fmt.Sprintf("%s:%s", r.lockeroptions.Prefix, key)
	unlockFunc = func() { r.Unlock(key) }
	if !r.IsLocked(key) {
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
				if msg.Pattern == eventChannel && lockKey == string(msg.Data) && !r.IsLocked(key) {
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
		r.Unlock(key)
		return unlockFunc, errors.New("timeout when waiting unlock another process")
	}
}

// NoopLocker

// IsLocked method
func (NoopLocker) IsLocked(string) bool { return false }

// IsLockedTTL method
func (NoopLocker) IsLockedTTL(string, time.Duration) bool { return false }

// HasBeenLocked method
func (NoopLocker) HasBeenLocked(string) bool { return false }

// Unlock method
func (NoopLocker) Unlock(string) {}

// Reset method
func (NoopLocker) Reset(string) {}

// Lock method
func (NoopLocker) Lock(string, time.Duration) (func(), error) { return func() {}, nil }

func (NoopLocker) Disconnect(context.Context) error { return nil }

// GetPrefix method
func (NoopLocker) GetPrefixLocker() string { return "" }

// GetTTLLocker method
func (NoopLocker) GetTTLLocker() time.Duration { return 0 }

// IsLockedTTLWithLimit method
func (NoopLocker) IsLockedTTLWithLimit(string, int, time.Duration) bool { return false }
