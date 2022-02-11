package database

import (
	"context"
	"github.com/golangid/candi/cache"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/gomodule/redigo/redis"
)

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
	deferFunc := logger.LogWithDefer("redis: disconnect...")
	defer deferFunc()

	if err := m.read.Close(); err != nil {
		return err
	}
	return m.write.Close()
}

// InitRedis connection from environment:
// REDIS_READ_DSN, REDIS_WRITE_DSN
func InitRedis() interfaces.RedisPool {
	deferFunc := logger.LogWithDefer("Load Redis connection...")
	defer deferFunc()

	inst := &redisInstance{
		read:  ConnectRedis(env.BaseEnv().DbRedisReadDSN),
		write: ConnectRedis(env.BaseEnv().DbRedisWriteDSN),
	}
	inst.cache = cache.NewRedisCache(inst.read, inst.write)
	return inst
}

// ConnectRedis connect to redis with dsn
func ConnectRedis(dsn string) *redis.Pool {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(dsn)
		},
	}

	ping := pool.Get()
	defer ping.Close()
	_, err := ping.Do("PING")
	if err != nil {
		panic("redis ping: " + err.Error())
	}

	return pool
}
