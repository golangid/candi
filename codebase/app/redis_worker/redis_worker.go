package redisworker

// Redis subscriber worker codebase

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candiutils"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

var (
	refreshWorkerNotif, shutdown, semaphore, startWorkerCh, releaseWorkerCh chan struct{}
)

type (
	redisWorker struct {
		ctx           context.Context
		ctxCancelFunc func()

		pubSubConn func() (subFn func() *redis.PubSubConn)
		isHaveJob  bool
		service    factory.ServiceFactory
		handlers   map[string]struct {
			handlerFunc   types.WorkerHandlerFunc
			errorHandlers []types.WorkerErrorHandler
		}
		consul *candiutils.Consul
		wg     sync.WaitGroup
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
				logger.LogYellow(fmt.Sprintf(`[REDIS-SUBSCRIBER] (key prefix): %-15s  --> (module): "%s"`, `"`+handler.Pattern+`"`, m.Name()))
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
	startWorkerCh, releaseWorkerCh = make(chan struct{}), make(chan struct{})

	workerInstance := &redisWorker{
		service:  service,
		handlers: handlers,
		pubSubConn: func() func() *redis.PubSubConn {
			conn := redisPool.Get()
			conn.Do("CONFIG", "SET", "notify-keyspace-events", "Ex")

			return func() *redis.PubSubConn {
				psc := &redis.PubSubConn{Conn: conn}
				psc.PSubscribe("__keyevent@*__:expired")
				return psc
			}
		},
		isHaveJob: len(handlers) != 0,
	}

	if env.BaseEnv().UseConsul {
		consul, err := candiutils.NewConsul(&candiutils.ConsulConfig{
			ConsulAgentHost:   env.BaseEnv().ConsulAgentHost,
			ConsulKey:         fmt.Sprintf("%s_redis_worker", service.Name()),
			LockRetryInterval: 1 * time.Second,
		})
		if err != nil {
			panic(err)
		}
		workerInstance.consul = consul
	}
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
				if r.consul != nil && count == env.BaseEnv().ConsulMaxJobRebalance {
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
	log.Println("\x1b[33;1mStopping Redis Subscriber worker...\x1b[0m")
	defer func() {
		if r.consul != nil {
			if err := r.consul.DestroySession(); err != nil {
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
	if r.consul == nil {
		go func() { startWorkerCh <- struct{}{} }()
		return
	}
	r.consul.DestroySession()
	hostname, _ := os.Hostname()
	value := map[string]string{
		"hostname": hostname,
	}
	go r.consul.RetryLockAcquire(value, startWorkerCh, releaseWorkerCh)
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
	trace, ctx := tracer.StartTraceWithContext(r.ctx, "RedisSubscriber")
	defer func() {
		if r := recover(); r != nil {
			tracer.SetError(ctx, fmt.Errorf("%v", r))
		}
		logger.LogGreen("redis_subscriber > trace_url: " + tracer.GetTraceURL(ctx))
		trace.Finish()
	}()

	if env.BaseEnv().DebugMode {
		log.Printf("\x1b[35;3mRedis Key Expired Subscriber: executing event key '%s'\x1b[0m", handlerName)
	}

	trace.SetTag("handler_name", handlerName)
	trace.SetTag("message", string(message))

	handler := r.handlers[handlerName]
	if err := handler.handlerFunc(ctx, message); err != nil {
		for _, errHandler := range handler.errorHandlers {
			errHandler(ctx, types.RedisSubscriber, handlerName, message, err)
		}
		tracer.SetError(ctx, err)
	}
}
