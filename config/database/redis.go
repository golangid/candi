package database

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"pkg.agungdwiprasetyo.com/candi/cache"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/config/env"
	"pkg.agungdwiprasetyo.com/candi/logger"
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
// REDIS_READ_DSN, REDIS_READ_TLS
// REDIS_WRITE_DSN, REDIS_WRITE_TLS
func InitRedis() interfaces.RedisPool {
	deferFunc := logger.LogWithDefer("Load Redis connection...")
	defer deferFunc()

	inst := new(redisInstance)

	inst.read = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(env.BaseEnv().DbRedisReadDSN)
		},
	}

	pingRead := inst.read.Get()
	defer pingRead.Close()
	_, err := pingRead.Do("PING")
	if err != nil {
		panic("redis read: " + err.Error())
	}

	inst.write = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(env.BaseEnv().DbRedisWriteDSN)
		},
	}

	pingWrite := inst.write.Get()
	defer pingWrite.Close()
	_, err = pingWrite.Do("PING")
	if err != nil {
		panic("redis write: " + err.Error())
	}

	inst.cache = cache.NewRedisCache(inst.read, inst.write)

	return inst
}
