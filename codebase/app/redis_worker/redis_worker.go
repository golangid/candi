package redisworker

// Redis subscriber worker codebase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gomodule/redigo/redis"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config/env"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

var (
	shutdown, semaphore chan struct{}
)

type (
	redisWorker struct {
		pubSubConn func() redis.PubSubConn
		isHaveJob  bool
		service    factory.ServiceFactory
		handlers   map[string]struct {
			handlerFunc   types.WorkerHandlerFunc
			errorHandlers []types.WorkerErrorHandler
		}
		wg sync.WaitGroup
	}

	handler struct {
		name    string
		message []byte
	}
)

// NewWorker create new redis subscriber
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	redisPool := service.GetDependency().GetRedisPool().WritePool()

	handlers := make(map[string]struct {
		handlerFunc   types.WorkerHandlerFunc
		errorHandlers []types.WorkerErrorHandler
	})
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.RedisSubscriber); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				logger.LogYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER] (key prefix): "%s" (processed by module): %s`, handler.Pattern, m.Name()))
				handlers[strings.Replace(handler.Pattern, "~", "", -1)] = struct {
					handlerFunc   types.WorkerHandlerFunc
					errorHandlers []types.WorkerErrorHandler
				}{
					handlerFunc: handler.HandlerFunc, errorHandlers: handler.ErrorHandler,
				}
			}
		}
	}

	if len(handlers) == 0 {
		log.Println("redis subscriber: no topic provided")
	} else {
		fmt.Printf("\x1b[34;1m⇨ Redis pubsub worker running with %d keys\x1b[0m\n\n", len(handlers))
	}

	shutdown, semaphore = make(chan struct{}, 1), make(chan struct{}, env.BaseEnv().MaxGoroutines)

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
		isHaveJob: len(handlers) != 0,
	}
}

func (r *redisWorker) Serve() {
	if !r.isHaveJob {
		return
	}

	psc := r.pubSubConn()

	// listen redis subscriber
	handlerReceiver := make(chan handler)
	go func() {
		for {
			switch msg := psc.Receive().(type) {
			case redis.Message:

				handlerName, messageData := candihelper.ParseRedisPubSubKeyTopic(string(msg.Data))
				_, ok := r.handlers[handlerName]
				if ok {
					handlerReceiver <- handler{
						name:    handlerName,
						message: []byte(messageData),
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
		case recv := <-handlerReceiver:

			r.wg.Add(1)
			semaphore <- struct{}{}
			go func(h handler) {
				defer r.wg.Done()

				trace := tracer.StartTrace(context.Background(), "RedisSubscriber")
				defer trace.Finish()
				ctx, tags := trace.Context(), trace.Tags()
				defer func() {
					if r := recover(); r != nil {
						tracer.SetError(ctx, fmt.Errorf("%v", r))
					}
					<-semaphore
					logger.LogGreen("redis subscriber " + tracer.GetTraceURL(ctx))
				}()

				if env.BaseEnv().DebugMode {
					log.Printf("\x1b[35;3mRedis Key Expired Subscriber: executing event key '%s'\x1b[0m", h.name)
				}

				tags["handler_name"] = h.name
				tags["message"] = string(h.message)

				handler := r.handlers[h.name]
				if err := handler.handlerFunc(ctx, recv.message); err != nil {
					for _, errHandler := range handler.errorHandlers {
						errHandler(ctx, types.RedisSubscriber, h.name, h.message, err)
					}
					tracer.SetError(ctx, err)
				}

			}(recv)

		case <-shutdown:
			return
		}
	}
}

func (r *redisWorker) Shutdown(ctx context.Context) {
	log.Println("\x1b[33;1mStopping Redis Subscriber worker...\x1b[0m")
	defer func() { log.Println("\x1b[33;1mStopping Redis Subscriber:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m") }()

	if !r.isHaveJob {
		return
	}

	shutdown <- struct{}{}
	runningJob := len(semaphore)
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mRedis Subscriber:\x1b[0m waiting %d job until done...\n", runningJob)
	}

	r.wg.Wait()
}
