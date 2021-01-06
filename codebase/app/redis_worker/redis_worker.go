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
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/candiutils"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config/env"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

var (
	refreshWorkerNotif, shutdown, semaphore, startWorkerCh, releaseWorkerCh chan struct{}
)

type (
	redisWorker struct {
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
	trace := tracer.StartTrace(context.Background(), "RedisSubscriber")
	defer trace.Finish()
	ctx, tags := trace.Context(), trace.Tags()
	defer func() {
		if r := recover(); r != nil {
			tracer.SetError(ctx, fmt.Errorf("%v", r))
		}
		logger.LogGreen("redis subscriber " + tracer.GetTraceURL(ctx))
	}()

	if env.BaseEnv().DebugMode {
		log.Printf("\x1b[35;3mRedis Key Expired Subscriber: executing event key '%s'\x1b[0m", handlerName)
	}

	tags["handler_name"] = handlerName
	tags["message"] = string(message)

	handler := r.handlers[handlerName]
	if err := handler.handlerFunc(ctx, message); err != nil {
		for _, errHandler := range handler.errorHandlers {
			errHandler(ctx, types.RedisSubscriber, handlerName, message, err)
		}
		tracer.SetError(ctx, err)
	}
}
