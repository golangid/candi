package redisworker

// Redis subscriber worker codebase

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/golangid/candi/broker"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/gomodule/redigo/redis"
)

var (
	refreshWorkerNotif, shutdown chan struct{}
)

type (
	redisWorker struct {
		ctx           context.Context
		ctxCancelFunc func()
		opt           option

		broker      interfaces.Broker
		isHaveJob   bool
		service     factory.ServiceFactory
		handlers    map[string]types.WorkerHandler
		wg          sync.WaitGroup
		semaphore   map[string]chan struct{}
		messagePool sync.Pool
	}
)

// NewWorker create new redis subscriber
func NewWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	workerInstance := &redisWorker{
		service:   service,
		opt:       getDefaultOption(),
		semaphore: make(map[string]chan struct{}),
		messagePool: sync.Pool{
			New: func() interface{} {
				return candishared.NewEventContext(bytes.NewBuffer(make([]byte, 0, 256)))
			},
		},
	}

	redisPool := service.GetDependency().GetRedisPool().WritePool()
	workerInstance.broker = service.GetDependency().GetBroker(types.RedisSubscriber)
	if workerInstance.broker == nil {
		workerInstance.broker = broker.NewRedisBroker(redisPool)
	}
	workerInstance.opt.locker = candiutils.NewRedisLocker(redisPool)

	for _, opt := range opts {
		opt(&workerInstance.opt)
	}
	workerInstance.opt.locker.Reset(fmt.Sprintf("%s:redis-worker-lock:*", service.Name()))

	handlers := make(map[string]types.WorkerHandler)
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.RedisSubscriber); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				logger.LogYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER] (key prefix): %-15s  --> (module): "%s"`, `"`+handler.Pattern+`"`, m.Name()))
				workerInstance.semaphore[handler.Pattern] = make(chan struct{}, workerInstance.opt.maxGoroutines)
				handlers[handler.Pattern] = handler
			}
		}
	}

	if len(handlers) == 0 {
		log.Println("redis subscriber: no topic provided")
	} else {
		fmt.Printf("\x1b[34;1m⇨ Redis pubsub worker running with %d keys\x1b[0m\n\n", len(handlers))
	}

	shutdown = make(chan struct{}, 1)

	workerInstance.handlers = handlers
	workerInstance.isHaveJob = len(handlers) != 0
	workerInstance.ctx, workerInstance.ctxCancelFunc = context.WithCancel(context.Background())

	return workerInstance
}

func (r *redisWorker) Serve() {
	if !r.isHaveJob {
		return
	}

	psc := r.broker.GetConfiguration().(*redis.PubSubConn)
	for {
		switch msg := psc.Receive().(type) {
		case redis.Message:
			redisMessage := broker.ParseRedisPubSubKeyTopic(msg.Data)
			if _, ok := r.handlers[redisMessage.HandlerName]; ok {
				r.semaphore[redisMessage.HandlerName] <- struct{}{}
				r.wg.Add(1)
				go func(message broker.RedisMessage) {
					defer func() {
						r.wg.Done()
						<-r.semaphore[message.HandlerName]
					}()

					if r.ctx.Err() != nil {
						logger.LogRed("redis_subscriber > ctx root err: " + r.ctx.Err().Error())
						return
					}
					r.processMessage(message)
				}(redisMessage)
			}

		case error:
			// if network connection error, create new connection from pool
			psc.Close()
			psc = r.broker.GetConfiguration().(*redis.PubSubConn)
		}
	}
}

func (r *redisWorker) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping Redis Subscriber:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	if !r.isHaveJob {
		return
	}

	shutdown <- struct{}{}
	runningJob := 0
	for _, sem := range r.semaphore {
		runningJob += len(sem)
	}
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mRedis Subscriber:\x1b[0m waiting %d job until done...\n", runningJob)
	}

	r.wg.Wait()
	r.ctxCancelFunc()
	r.opt.locker.Reset(fmt.Sprintf("%s:redis-worker-lock:*", r.service.Name()))
}

func (r *redisWorker) Name() string {
	return string(types.RedisSubscriber)
}

func (r *redisWorker) processMessage(param broker.RedisMessage) {
	// lock for multiple worker (if running on multiple pods/instance)
	if r.opt.locker.IsLocked(r.getLockKey(param.HandlerName, param.EventID)) {
		logger.LogYellow("redis_subscriber > eventID " + param.EventID + " is locked")
		return
	}
	defer r.opt.locker.Unlock(r.getLockKey(param.HandlerName, param.EventID))

	ctx := r.ctx
	selectedHandler := r.handlers[param.HandlerName]
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "RedisSubscriber", map[string]string{})
	defer func() {
		if r := recover(); r != nil {
			trace.SetError(fmt.Errorf("panic: %v", r))
			tracer.LogStackTrace(trace)
		}
		logger.LogGreen("redis_subscriber > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish()
	}()

	if r.opt.debugMode {
		log.Printf("\x1b[35;3mRedis Key Expired Subscriber: executing event key '%s'\x1b[0m", param.HandlerName)
	}

	trace.SetTag("handler_name", param.HandlerName)
	trace.SetTag("event_id", param.EventID)
	trace.Log("message", param.Message)

	eventContext := r.messagePool.Get().(*candishared.EventContext)
	defer r.releaseMessagePool(eventContext)
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(types.RedisSubscriber))
	eventContext.SetHandlerRoute(param.HandlerName)
	eventContext.SetKey(param.EventID)
	eventContext.WriteString(param.Message)

	for _, handlerFunc := range selectedHandler.HandlerFuncs {
		if err := handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
			trace.SetError(err)
		}
	}
}

func (r *redisWorker) getLockKey(handlerName, eventID string) string {
	return fmt.Sprintf("%s:redis-worker-lock:%s-%s", r.service.Name(), handlerName, eventID)
}

func (r *redisWorker) releaseMessagePool(eventContext *candishared.EventContext) {
	eventContext.Reset()
	r.messagePool.Put(eventContext)
}
