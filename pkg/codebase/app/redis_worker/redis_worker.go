package redisworker

// Redis subscriber worker codebase

import (
	"context"
	"fmt"
	"log"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"github.com/gomodule/redigo/redis"
)

type redisWorker struct {
	pubSubConn func() redis.PubSubConn
	service    factory.ServiceFactory
}

// NewWorker create new redis subscriber
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	redisPool := service.GetDependency().GetRedisPool().WritePool()

	return &redisWorker{
		service: service,
		pubSubConn: func() redis.PubSubConn {
			conn := redisPool.Get()
			conn.Do("CONFIG", "SET", "notify-keyspace-events", "Ex")

			psc := redis.PubSubConn{Conn: conn}
			psc.PSubscribe("__keyevent@*__:expired")

			return psc
		},
	}
}

func (r *redisWorker) Serve() {
	handlers := make(map[string]interfaces.WorkerHandler)
	for _, m := range r.service.GetModules() {
		if h := m.WorkerHandler(constant.RedisSubscriber); h != nil {
			for _, topic := range h.GetTopics() {
				handlers[topic] = h
			}
		}
	}

	psc := r.pubSubConn()

	fmt.Printf("\x1b[34;1m⇨ Redis pubsub worker running\x1b[0m\n\n")
	for {
		switch msg := psc.Receive().(type) {
		case redis.Message:
			func() {
				defer func() { recover() }()
				modName, handlerName, messageData := helper.ParseRedisPubSubKeyTopic(string(msg.Data))
				handlers[modName].ProcessMessage(context.Background(), handlerName, []byte(messageData))
			}()
		case error:

			// if network connection error, create new connection from pool
			psc = r.pubSubConn()
		}
	}
}

func (r *redisWorker) Shutdown(ctx context.Context) {
	log.Println("Stopping redis subscriber worker...")
	// TODO: handling graceful stop all channel from redis subscriber
}
