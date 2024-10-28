package redisworker

// Redis subscriber worker codebase

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/golangid/candi/broker"
	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/gomodule/redigo/redis"
)

type (
	redisWorker struct {
		ctx           context.Context
		ctxCancelFunc func()
		opt           option

		bk          *broker.RedisBroker
		isHaveJob   bool
		service     factory.ServiceFactory
		handlers    map[string]types.WorkerHandler
		wg          sync.WaitGroup
		semaphore   map[string]chan struct{}
		messagePool sync.Pool
	}
)

// NewWorker create new redis subscriber
func NewWorker(service factory.ServiceFactory, bk interfaces.Broker, opts ...OptionFunc) factory.AppServerFactory {
	workerInstance := &redisWorker{
		service:   service,
		opt:       getDefaultOption(),
		semaphore: make(map[string]chan struct{}),
		messagePool: sync.Pool{
			New: func() any {
				return candishared.NewEventContext(bytes.NewBuffer(make([]byte, 0, 256)))
			},
		},
	}

	redisPool := service.GetDependency().GetRedisPool().WritePool()
	workerInstance.bk, _ = bk.(*broker.RedisBroker)
	if workerInstance.bk == nil {
		workerInstance.bk = broker.NewRedisBroker(redisPool)
	}
	workerInstance.opt.locker = candiutils.NewRedisLocker(redisPool)

	for _, opt := range opts {
		opt(&workerInstance.opt)
	}
	workerInstance.opt.locker.Reset(fmt.Sprintf("%s:redis-worker-lock:*", service.Name()))

	handlers := make(map[string]types.WorkerHandler)
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(workerInstance.bk.WorkerType); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				logger.LogYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER]%s (key prefix): %-15s  --> (module): "%s"`, getWorkerTypeLog(workerInstance.bk.WorkerType), `"`+handler.Pattern+`"`, m.Name()))
				workerInstance.semaphore[handler.Pattern] = make(chan struct{}, workerInstance.opt.maxGoroutines)
				handlers[handler.Pattern] = handler
			}
		}
	}

	if len(handlers) == 0 {
		log.Println("redis subscriber: no topic provided")
	} else {
		fmt.Printf("\x1b[34;1m⇨ Redis pubsub worker%s running with %d keys\x1b[0m\n\n", getWorkerTypeLog(workerInstance.bk.WorkerType), len(handlers))
	}

	workerInstance.handlers = handlers
	workerInstance.isHaveJob = len(handlers) != 0
	workerInstance.ctx, workerInstance.ctxCancelFunc = context.WithCancel(context.Background())

	return workerInstance
}

func (r *redisWorker) Serve() {
	if !r.isHaveJob {
		return
	}

	psc := r.bk.InitPubSubConn()
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
			psc = r.bk.InitPubSubConn()
		}
	}
}

func (r *redisWorker) Shutdown(ctx context.Context) {
	defer func() {
		fmt.Printf("\r%s \x1b[33;1mStopping Redis Subscriber%s:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m%s\n",
			time.Now().Format(candihelper.TimeFormatLogger), getWorkerTypeLog(r.bk.WorkerType), strings.Repeat(" ", 20))
	}()

	if !r.isHaveJob {
		return
	}

	runningJob := 0
	for _, sem := range r.semaphore {
		runningJob += len(sem)
	}
	waitingJob := "... "
	if runningJob != 0 {
		waitingJob = fmt.Sprintf("waiting %d job until done... ", runningJob)
	}
	fmt.Printf("\r%s \x1b[33;1mStopping Redis Subscriber%s:\x1b[0m %s",
		time.Now().Format(candihelper.TimeFormatLogger), getWorkerTypeLog(r.bk.WorkerType), waitingJob)

	r.wg.Wait()
	r.ctxCancelFunc()
	r.opt.locker.Reset(fmt.Sprintf("%s:redis-worker-lock:*", r.service.Name()))
}

func (r *redisWorker) Name() string {
	return string(r.bk.WorkerType)
}

func (r *redisWorker) processMessage(param broker.RedisMessage) {
	// lock for multiple worker (if running on multiple pods/instance)
	lockKey := r.getLockKey(param.HandlerName, param.EventID)
	if r.opt.locker.IsLocked(lockKey) {
		logger.LogYellow("redis_subscriber > eventID " + param.EventID + " is locked")
		return
	}
	defer r.opt.locker.Unlock(lockKey)

	var err error
	message := []byte(param.Message)
	if len(message) == 0 {
		conn := r.bk.Pool.Get()
		message, err = redis.Bytes(conn.Do("HGET", broker.RedisBrokerKey, param.Key))
		if err != nil {
			return
		}
		conn.Do("HDEL", broker.RedisBrokerKey, param.Key)
		conn.Close()
	}

	ctx := r.ctx
	selectedHandler := r.handlers[param.HandlerName]
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "RedisSubscriber", make(map[string]string, 0))
	defer func() {
		if r := recover(); r != nil {
			trace.SetTag("panic", true)
			err = fmt.Errorf("%v", r)
		}
		trace.Finish(tracer.FinishWithError(err))
	}()

	if r.opt.debugMode {
		log.Printf("\x1b[35;3mRedis Key Expired Subscriber%s: executing event topic '%s'\x1b[0m", getWorkerTypeLog(r.bk.WorkerType), param.HandlerName)
	}

	trace.SetTag("handler_name", param.HandlerName)
	trace.SetTag("event_id", param.EventID)
	if r.bk.WorkerType != types.RedisSubscriber {
		trace.SetTag("worker_type", string(r.bk.WorkerType))
	}
	trace.Log("message", message)

	eventContext := r.messagePool.Get().(*candishared.EventContext)
	defer r.releaseMessagePool(eventContext)
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(r.bk.WorkerType))
	eventContext.SetHandlerRoute(param.HandlerName)
	eventContext.SetKey(param.Key)
	eventContext.SetHeader(map[string]string{
		"event_id": param.EventID,
	})
	eventContext.Write(message)

	for _, handlerFunc := range selectedHandler.HandlerFuncs {
		if err = handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
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

func getWorkerTypeLog(name types.Worker) (workerType string) {
	if name != types.RedisSubscriber {
		workerType = " [worker_type: " + string(name) + "]"
	}
	return
}
