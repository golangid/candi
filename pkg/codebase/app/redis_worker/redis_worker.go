package redisworker

// Redis subscriber worker codebase

import (
	"context"
	"fmt"
	"sync"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"github.com/gomodule/redigo/redis"
)

type redisWorker struct {
	pubSubConn func() redis.PubSubConn
	service    factory.ServiceFactory
	shutdown   chan struct{}
	wg         sync.WaitGroup
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
		shutdown: make(chan struct{}),
	}
}

func (r *redisWorker) Serve() {
	handlers := make(map[string]interfaces.WorkerHandler)
	for _, m := range r.service.GetModules() {
		if h := m.WorkerHandler(constant.RedisSubscriber); h != nil {
			for _, topic := range h.GetTopics() {
				fmt.Println(helper.StringYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER] (key prefix): "%-10s"  (processed by module): %s`, topic, m.Name())))
				handlers[topic] = h
			}
		}
	}

	psc := r.pubSubConn()

	// listen redis subscriber
	messageReceiver := make(chan []byte)
	go func() {
		for {
			switch msg := psc.Receive().(type) {
			case redis.Message:
				messageReceiver <- msg.Data
			case error:
				// if network connection error, create new connection from pool
				psc = r.pubSubConn()
			}
		}
	}()

	// run worker with listen shutdown channel
	fmt.Printf("\x1b[34;1m⇨ Redis pubsub worker running with %d keys\x1b[0m\n\n", len(handlers))
	for {
		select {
		case message := <-messageReceiver:
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				defer func() { recover() }()

				modName, handlerName, messageData := helper.ParseRedisPubSubKeyTopic(string(message))
				handlers[modName].ProcessMessage(context.Background(), handlerName, []byte(messageData))
			}()
		case <-r.shutdown:
			break
		}
	}
}

func (r *redisWorker) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping redis subscriber worker...")
	defer deferFunc()

	r.shutdown <- struct{}{}

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		fmt.Print("redis-subscriber: force shutdown ")
	case <-done:
		fmt.Print("redis-subscriber: success waiting all job until done ")
	}
}
