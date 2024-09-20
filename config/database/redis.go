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

type RedisInstance struct {
	DBRead, DBWrite *redis.Pool
	ICache          interfaces.Cache
}

func (m *RedisInstance) ReadPool() *redis.Pool {
	return m.DBRead
}

func (m *RedisInstance) WritePool() *redis.Pool {
	return m.DBWrite
}

func (m *RedisInstance) Health() map[string]error {
	mErr := make(map[string]error)
	if m.DBWrite != nil {
		connWrite := m.DBWrite.Get()
		defer connWrite.Close()
		_, err := connWrite.Do("PING")
		mErr["redis_write"] = err
	}
	if m.DBRead != nil {
		connRead := m.DBWrite.Get()
		defer connRead.Close()
		_, err := connRead.Do("PING")
		mErr["redis_read"] = err
	}
	return mErr
}

func (m *RedisInstance) Cache() interfaces.Cache {
	return m.ICache
}

func (m *RedisInstance) Disconnect(ctx context.Context) (err error) {
	defer logger.LogWithDefer("\x1b[33;5mredis\x1b[0m: disconnect...")()

	if m.DBRead != nil {
		if err := m.DBRead.Close(); err != nil {
			return err
		}
	}
	if m.DBWrite != nil {
		err = m.DBWrite.Close()
	}
	return
}

// InitRedis connection from environment:
// REDIS_READ_DSN, REDIS_WRITE_DSN
// if want to create single connection, use REDIS_WRITE_DSN and set empty for REDIS_READ_DSN
func InitRedis(opts ...RedisPoolOption) *RedisInstance {
	defer logger.LogWithDefer("Load Redis connection...")()

	connReadDSN, connWriteDSN := env.BaseEnv().DbRedisReadDSN, env.BaseEnv().DbRedisWriteDSN
	if connReadDSN == "" {
		poolConn := ConnectRedis(connWriteDSN, opts...)
		return &RedisInstance{
			DBRead:  poolConn,
			DBWrite: poolConn,
			ICache:  cache.NewRedisCache(poolConn, poolConn),
		}
	}

	inst := &RedisInstance{
		DBRead:  ConnectRedis(connReadDSN, opts...),
		DBWrite: ConnectRedis(connWriteDSN, opts...),
	}
	inst.ICache = cache.NewRedisCache(inst.DBRead, inst.DBWrite)
	return inst
}

// ConnectRedis connect to redis with dsn
func ConnectRedis(dsn string, opts ...RedisPoolOption) *redis.Pool {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(dsn)
		},

		// default pool config
		MaxIdle:         50,
		MaxActive:       80,
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
