package redis

import (
	"context"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"github.com/gomodule/redigo/redis"
)

type RedisRepo struct {
	pool *redis.Pool
}

func NewRedisRepo(pool *redis.Pool) *RedisRepo {
	return &RedisRepo{
		pool: pool,
	}
}

func (r *RedisRepo) SaveScheduledNotification(ctx context.Context, key string, data []byte, exp time.Duration) error {
	conn := r.pool.Get()
	defer conn.Close()

	key += ":" + string(data)

	_, err := conn.Do("SETEX", key, int(exp.Seconds()), "")
	if err != nil {
		logger.LogE(err.Error())
	}
	return err
}
