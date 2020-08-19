package redisworker

// Redis subscriber worker codebase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
	"github.com/gomodule/redigo/redis"
)

type redisWorker struct {
	pubSubConn func() redis.PubSubConn
	isHaveJob  bool
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
	handlers := make(map[string]types.WorkerHandlerFunc)
	for _, m := range r.service.GetModules() {
		if h := m.WorkerHandler(types.RedisSubscriber); h != nil {
			for topic, handlerFunc := range h.MountHandlers() {
				logger.LogYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER] (key prefix): "%s" (processed by module): %s`, topic, m.Name()))
				handlers[strings.Replace(topic, "~", "", -1)] = handlerFunc
			}
		}
	}

	if len(handlers) == 0 {
		log.Println("redis subscriber: no topic provided")
		return
	}
	r.isHaveJob = true

	psc := r.pubSubConn()

	// listen redis subscriber
	handlerReceiver := make(chan struct {
		handlerName string
		message     []byte
	})
	go func() {
		for {
			switch msg := psc.Receive().(type) {
			case redis.Message:

				handlerName, messageData := helper.ParseRedisPubSubKeyTopic(string(msg.Data))
				_, ok := handlers[handlerName]
				if ok {
					handlerReceiver <- struct {
						handlerName string
						message     []byte
					}{
						handlerName: handlerName,
						message:     []byte(messageData),
					}
				}

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
		case handler := <-handlerReceiver:
			r.wg.Add(1)
			go tracer.WithTraceFunc(context.Background(), "RedisSubscriber", func(ctx context.Context, tags map[string]interface{}) {
				defer r.wg.Done()
				defer func() {
					if r := recover(); r != nil {
						tracer.SetError(ctx, fmt.Errorf("%v", r))
					}
					logger.LogGreen(tracer.GetTraceURL(ctx))
				}()

				tags["handler_name"] = handler.handlerName
				tags["message"] = string(handler.message)

				handlerFunc := handlers[handler.handlerName]
				if err := handlerFunc(ctx, handler.message); err != nil {
					panic(err)
				}
			})

		case <-r.shutdown:
			return
		}
	}
}

func (r *redisWorker) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping redis subscriber worker...")
	defer deferFunc()

	if !r.isHaveJob {
		return
	}

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
