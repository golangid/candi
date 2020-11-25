package database

import (
	"context"
	"fmt"
	"strconv"

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
// REDIS_READ_HOST, REDIS_READ_PORT, REDIS_READ_AUTH, REDIS_READ_TLS, REDIS_READ_DB,
// REDIS_WRITE_HOST, REDIS_WRITE_PORT, REDIS_WRITE_AUTH, REDIS_WRITE_TLS, REDIS_WRITE_DB
func InitRedis() interfaces.RedisPool {
	deferFunc := logger.LogWithDefer("Load Redis connection...")
	defer deferFunc()

	inst := new(redisInstance)

	inst.read = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			redisDB, _ := strconv.Atoi(env.BaseEnv().DbRedisReadDBIndex)
			return redis.Dial("tcp", fmt.Sprintf("%s:%s", env.BaseEnv().DbRedisReadHost, env.BaseEnv().DbRedisReadPort),
				redis.DialPassword(env.BaseEnv().DbRedisReadAuth),
				redis.DialDatabase(redisDB),
				redis.DialUseTLS(env.BaseEnv().DbRedisReadTLS))
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
			redisDB, _ := strconv.Atoi(env.BaseEnv().DbRedisWriteDBIndex)
			return redis.Dial("tcp", fmt.Sprintf("%s:%s", env.BaseEnv().DbRedisWriteHost, env.BaseEnv().DbRedisWritePort),
				redis.DialPassword(env.BaseEnv().DbRedisWriteAuth),
				redis.DialDatabase(redisDB),
				redis.DialUseTLS(env.BaseEnv().DbRedisWriteTLS))
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
