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
	handlers   map[string]types.WorkerHandlerFunc
	shutdown   chan struct{}
	wg         sync.WaitGroup
}

// NewWorker create new redis subscriber
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	redisPool := service.GetDependency().GetRedisPool().WritePool()

	handlers := make(map[string]types.WorkerHandlerFunc)
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.RedisSubscriber); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				logger.LogYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER] (key prefix): "%s" (processed by module): %s`, handler.Pattern, m.Name()))
				handlers[strings.Replace(handler.Pattern, "~", "", -1)] = handler.HandlerFunc
			}
		}
	}

	if len(handlers) == 0 {
		log.Println("redis subscriber: no topic provided")
	} else {
		fmt.Printf("\x1b[34;1m⇨ Redis pubsub worker running with %d keys\x1b[0m\n\n", len(handlers))
	}

	return &redisWorker{
		service:  service,
		handlers: handlers,
		pubSubConn: func() redis.PubSubConn {
			conn := redisPool.Get()
			conn.Do("CONFIG", "SET", "notify-keyspace-events", "Ex")

			psc := redis.PubSubConn{Conn: conn}
			psc.PSubscribe("__keyevent@*__:expired")

			return psc
		},
		shutdown:  make(chan struct{}),
		isHaveJob: len(handlers) != 0,
	}
}

func (r *redisWorker) Serve() {
	if !r.isHaveJob {
		return
	}

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
				_, ok := r.handlers[handlerName]
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

				log.Printf("\x1b[35;3mRedis Key Expired Subscriber: executing event key '%s'\x1b[0m", handler.handlerName)

				handlerFunc := r.handlers[handler.handlerName]
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
