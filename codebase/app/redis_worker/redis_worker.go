package redisworker

// Redis subscriber worker codebase

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
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

		pool      *redis.Pool
		isHaveJob bool
		service   factory.ServiceFactory
		handlers  map[string]types.WorkerHandler
		wg        sync.WaitGroup
		semaphore map[string]chan struct{}
	}
)

// NewWorker create new redis subscriber
func NewWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	redisPool := service.GetDependency().GetRedisPool().WritePool()

	workerInstance := &redisWorker{
		service:   service,
		opt:       getDefaultOption(),
		semaphore: make(map[string]chan struct{}),
		pool:      redisPool,
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

	psc := r.getPubSubConn()
	messageChan := make(chan redis.Message)

	go func() {
		for {
			switch msg := psc.Receive().(type) {
			case redis.Message:
				messageChan <- msg

			case error:
				// if network connection error, create new connection from pool
				psc.Close()
				psc = r.getPubSubConn()
			}
		}
	}()

	for {
		select {
		case <-shutdown:
			psc.Unsubscribe()
			return

		case msg := <-messageChan:
			redisMessage := ParseRedisPubSubKeyTopic(msg.Data)
			if _, ok := r.handlers[redisMessage.HandlerName]; ok {

				r.semaphore[redisMessage.HandlerName] <- struct{}{}
				r.wg.Add(1)
				go func(message RedisMessage) {
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
}

func (r *redisWorker) Name() string {
	return string(types.RedisSubscriber)
}

func (r *redisWorker) getPubSubConn() *redis.PubSubConn {
	conn := r.pool.Get()
	conn.Do("CONFIG", "SET", "notify-keyspace-events", "Ex")

	psc := &redis.PubSubConn{Conn: conn}
	psc.PSubscribe("__keyevent@*__:expired")
	return psc
}

func (r *redisWorker) processMessage(param RedisMessage) {

	// lock for multiple worker (if running on multiple pods/instance)
	if r.opt.locker.IsLocked(r.getLockKey(param.HandlerName, param.EventID)) {
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
			trace.Log("stacktrace", string(debug.Stack()))
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

	var eventContext candishared.EventContext
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(types.RedisSubscriber))
	eventContext.SetHandlerRoute(param.HandlerName)
	eventContext.SetKey(param.EventID)
	eventContext.WriteString(param.Message)

	for _, handlerFunc := range selectedHandler.HandlerFuncs {
		if err := handlerFunc(&eventContext); err != nil {
			eventContext.SetError(err)
			trace.SetError(err)
		}
	}
}

func (r *redisWorker) getLockKey(handlerName, eventID string) string {
	return fmt.Sprintf("%s:redis-worker-lock:%s-%s", r.service.Name(), handlerName, eventID)
}
