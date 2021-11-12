package redisworker

// Redis subscriber worker codebase

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/gomodule/redigo/redis"
)

var (
	refreshWorkerNotif, shutdown, semaphore, startWorkerCh, releaseWorkerCh chan struct{}
)

type (
	redisWorker struct {
		ctx           context.Context
		ctxCancelFunc func()
		opt           option

		pubSubConn func() (subFn func() *redis.PubSubConn)
		isHaveJob  bool
		service    factory.ServiceFactory
		handlers   map[string]types.WorkerHandler
		wg         sync.WaitGroup
	}
)

// NewWorker create new redis subscriber
func NewWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	redisPool := service.GetDependency().GetRedisPool().WritePool()

	workerInstance := &redisWorker{
		service: service,
		opt:     getDefaultOption(),
	}
	workerInstance.opt.locker = candiutils.NewRedisLocker(redisPool.Get())

	for _, opt := range opts {
		opt(&workerInstance.opt)
	}

	handlers := make(map[string]types.WorkerHandler)
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.RedisSubscriber); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				logger.LogYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER] (key prefix): %-15s  --> (module): "%s"`, `"`+handler.Pattern+`"`, m.Name()))
				handlers[strings.Replace(handler.Pattern, "~", "", -1)] = handler
			}
		}
	}

	if len(handlers) == 0 {
		log.Println("redis subscriber: no topic provided")
	} else {
		fmt.Printf("\x1b[34;1m⇨ Redis pubsub worker running with %d keys\x1b[0m\n\n", len(handlers))
	}

	shutdown, semaphore = make(chan struct{}, 1), make(chan struct{}, workerInstance.opt.maxGoroutines)
	startWorkerCh, releaseWorkerCh = make(chan struct{}), make(chan struct{})

	workerInstance.handlers = handlers
	workerInstance.pubSubConn = func() func() *redis.PubSubConn {
		conn := redisPool.Get()
		conn.Do("CONFIG", "SET", "notify-keyspace-events", "Ex")

		return func() *redis.PubSubConn {
			psc := &redis.PubSubConn{Conn: conn}
			psc.PSubscribe("__keyevent@*__:expired")
			return psc
		}
	}
	workerInstance.isHaveJob = len(handlers) != 0
	workerInstance.ctx, workerInstance.ctxCancelFunc = context.WithCancel(context.Background())

	return workerInstance
}

func (r *redisWorker) Serve() {
	if !r.isHaveJob {
		return
	}

	r.createConsulSession()
	subFunc := r.pubSubConn()

START:
	select {
	case <-startWorkerCh:
		psc := subFunc()

		stopListener := make(chan struct{})
		countJobs := make(chan int)
		go r.runListener(stopListener, countJobs, psc)

		for {
			select {
			case count := <-countJobs:
				if r.opt.consul != nil && count == r.opt.consul.MaxJobRebalance {
					// recreate session
					r.createConsulSession()
					<-releaseWorkerCh
					psc.PUnsubscribe()
					go func() { stopListener <- struct{}{} }()
					goto START
				}

			case <-shutdown:
				go func() { stopListener <- struct{}{} }()
				return
			}
		}

	case <-shutdown:
		return
	}
}

func (r *redisWorker) Shutdown(ctx context.Context) {
	defer func() {
		if r.opt.consul != nil {
			if err := r.opt.consul.DestroySession(); err != nil {
				panic(err)
			}
		}
		log.Println("\x1b[33;1mStopping Redis Subscriber:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")
	}()

	if !r.isHaveJob {
		return
	}

	shutdown <- struct{}{}
	runningJob := len(semaphore)
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mRedis Subscriber:\x1b[0m waiting %d job until done...\n", runningJob)
	}

	r.wg.Wait()
	r.ctxCancelFunc()
}

func (r *redisWorker) Name() string {
	return string(types.RedisSubscriber)
}

func (r *redisWorker) createConsulSession() {
	if r.opt.consul == nil {
		go func() { startWorkerCh <- struct{}{} }()
		return
	}
	r.opt.consul.DestroySession()
	hostname, _ := os.Hostname()
	value := map[string]string{
		"hostname": hostname,
	}
	go r.opt.consul.RetryLockAcquire(value, startWorkerCh, releaseWorkerCh)
}

func (r *redisWorker) runListener(stop <-chan struct{}, count chan<- int, psc *redis.PubSubConn) {
	defer func() {
		if r := recover(); r != nil {
			logger.LogE(fmt.Sprint(r))
		}
	}()

	totalRunJobs := 0
	// listen redis subscriber
	for {
		select {
		case <-stop:
			return

		default:
			switch msg := psc.Receive().(type) {
			case redis.Message:

				if r.opt.locker.IsLocked(fmt.Sprintf("redis-worker-lock-%s-%s", r.service.Name(), msg.Data)) {
					continue
				}

				handlerName, messageData := candihelper.ParseRedisPubSubKeyTopic(string(msg.Data))
				if _, ok := r.handlers[handlerName]; ok {

					semaphore <- struct{}{}
					r.wg.Add(1)
					go func(handlerName string, message []byte) {
						defer func() {
							r.wg.Done()
							<-semaphore
							totalRunJobs++
							count <- totalRunJobs
						}()

						if r.ctx.Err() != nil {
							logger.LogRed("redis_subscriber > ctx root err: " + r.ctx.Err().Error())
							return
						}
						r.processMessage(handlerName, message)
					}(handlerName, []byte(messageData))

				}

			case error:
				psc.Close()
				// if network connection error, create new connection from pool
				subFn := r.pubSubConn()
				psc = subFn()
			}
		}
	}
}

func (r *redisWorker) processMessage(handlerName string, message []byte) {
	ctx := r.ctx
	selectedHandler := r.handlers[handlerName]
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	trace, ctx := tracer.StartTraceWithContext(ctx, "RedisSubscriber")
	defer func() {
		if r := recover(); r != nil {
			tracer.SetError(ctx, fmt.Errorf("%v", r))
		}
		logger.LogGreen("redis_subscriber > trace_url: " + tracer.GetTraceURL(ctx))
		trace.Finish()
	}()

	if r.opt.debugMode {
		log.Printf("\x1b[35;3mRedis Key Expired Subscriber: executing event key '%s'\x1b[0m", handlerName)
	}

	trace.SetTag("handler_name", handlerName)
	trace.Log("message", string(message))

	if err := selectedHandler.HandlerFunc(ctx, message); err != nil {
		if selectedHandler.ErrorHandler != nil {
			selectedHandler.ErrorHandler(ctx, types.RedisSubscriber, handlerName, message, err)
		}
		tracer.SetError(ctx, err)
	}
}
