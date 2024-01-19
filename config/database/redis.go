package database

import (
	"context"
	"log"
	"time"

	"github.com/golangid/candi/cache"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/gomodule/redigo/redis"
)

type RedisPoolOption func(pool *redis.Pool)

func RedisPoolOptionSetMaxIdle(count int) RedisPoolOption {
	return func(pool *redis.Pool) { pool.MaxIdle = count }
}
func RedisPoolOptionSetMaxActive(count int) RedisPoolOption {
	return func(pool *redis.Pool) { pool.MaxActive = count }
}
func RedisPoolOptionSetIdleTimeout(dur time.Duration) RedisPoolOption {
	return func(pool *redis.Pool) { pool.IdleTimeout = dur }
}
func RedisPoolOptionSetMaxConnLifetime(dur time.Duration) RedisPoolOption {
	return func(pool *redis.Pool) { pool.MaxConnLifetime = dur }
}

type redisInstance struct {
	read, write *redis.Pool
	cache       interfaces.Cache
}

func (m *redisInstance) ReadPool() *redis.Pool {
	return m.read
}
func (m *redisInstance) WritePool() *redis.Pool {
	return m.write
}
func (m *redisInstance) Health() map[string]error {
	mErr := make(map[string]error)

	connWrite := m.write.Get()
	defer connWrite.Close()
	_, err := connWrite.Do("PING")
	mErr["redis_write"] = err

	connRead := m.write.Get()
	defer connRead.Close()
	_, err = connRead.Do("PING")
	mErr["redis_read"] = err

	return mErr
}
func (m *redisInstance) Cache() interfaces.Cache {
	return m.cache
}
func (m *redisInstance) Disconnect(ctx context.Context) (err error) {
	defer logger.LogWithDefer("redis: disconnect...")()

	if err := m.read.Close(); err != nil {
		return err
	}
	return m.write.Close()
}

// InitRedis connection from environment:
// REDIS_READ_DSN, REDIS_WRITE_DSN
func InitRedis(opts ...RedisPoolOption) interfaces.RedisPool {
	defer logger.LogWithDefer("Load Redis connection...")()

	inst := &redisInstance{
		read:  ConnectRedis(env.BaseEnv().DbRedisReadDSN, opts...),
		write: ConnectRedis(env.BaseEnv().DbRedisWriteDSN, opts...),
	}
	inst.cache = cache.NewRedisCache(inst.read, inst.write)
	return inst
}

// ConnectRedis connect to redis with dsn
func ConnectRedis(dsn string, opts ...RedisPoolOption) *redis.Pool {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(dsn)
		},

		// default pool config
		MaxIdle:         80,
		MaxActive:       500,
		IdleTimeout:     20 * time.Minute,
		MaxConnLifetime: 1 * time.Hour,
	}
	for _, opt := range opts {
		opt(pool)
	}

	ping := pool.Get()
	defer ping.Close()
	_, err := ping.Do("PING")
	if err != nil {
		log.Panicf("redis ping: %s", err.Error())
	}

	return pool
}
