package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	cronworker "pkg.agungdwiprasetyo.com/candi/codebase/app/cron_worker"
	graphqlserver "pkg.agungdwiprasetyo.com/candi/codebase/app/graphql_server"
	grpcserver "pkg.agungdwiprasetyo.com/candi/codebase/app/grpc_server"
	kafkaworker "pkg.agungdwiprasetyo.com/candi/codebase/app/kafka_worker"
	redisworker "pkg.agungdwiprasetyo.com/candi/codebase/app/redis_worker"
	restserver "pkg.agungdwiprasetyo.com/candi/codebase/app/rest_server"
	taskqueueworker "pkg.agungdwiprasetyo.com/candi/codebase/app/task_queue_worker"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

// App service
type App struct {
	servers []factory.AppServerFactory
}

// New service app
// New service app
func New(service factory.ServiceFactory) *App {
	log.Printf("Starting \x1b[32;1m%s\x1b[0m service\n\n", service.Name())

	// init service name tracer
	serviceName := string(service.Name())
	if config.BaseEnv().Environment != "" {
		serviceName = fmt.Sprintf("%s-%s", serviceName, strings.ToLower(config.BaseEnv().Environment))
	}
	// init tracer
	tracer.InitOpenTracing(config.BaseEnv().JaegerTracingHost, serviceName)
	// init logger
	logger.InitZap()

	appInstance := new(App)
	if config.BaseEnv().UseREST {
		appInstance.servers = append(appInstance.servers, restserver.NewServer(service))
	}

	if config.BaseEnv().UseGRPC {
		appInstance.servers = append(appInstance.servers, grpcserver.NewServer(service))
	}

	if !config.BaseEnv().UseREST && config.BaseEnv().UseGraphQL {
		appInstance.servers = append(appInstance.servers, graphqlserver.NewServer(service))
	}

	if config.BaseEnv().UseKafkaConsumer {
		appInstance.servers = append(appInstance.servers, kafkaworker.NewWorker(service))
	}

	if config.BaseEnv().UseCronScheduler {
		appInstance.servers = append(appInstance.servers, cronworker.NewWorker(service))
	}

	if config.BaseEnv().UseRedisSubscriber {
		appInstance.servers = append(appInstance.servers, redisworker.NewWorker(service))
	}

	if config.BaseEnv().UseTaskQueueWorker {
		appInstance.servers = append(appInstance.servers, taskqueueworker.NewWorker(service))
	}

	return appInstance
}

// Run start app
func (a *App) Run() {

	if len(a.servers) == 0 {
		panic("No server/worker running")
	}

	errServe := make(chan error)
	for _, server := range a.servers {
		go func(srv factory.AppServerFactory) {
			defer func() {
				if r := recover(); r != nil {
					errServe <- fmt.Errorf("%v", r)
				}
			}()
			srv.Serve()
		}(server)
	}

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case e := <-errServe:
		panic(e)
	case <-quitSignal:
		a.shutdown(quitSignal)
	}
}

// graceful shutdown all server, panic if there is still a process running when the request exceed given timeout in context
func (a *App) shutdown(forceShutdown chan os.Signal) {
	fmt.Println("\x1b[34;1mGracefully shutdown... (press Ctrl+C again to force)\x1b[0m")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)

		var wg sync.WaitGroup
		for _, server := range a.servers {
			wg.Add(1)
			go func(srv factory.AppServerFactory) {
				defer wg.Done()
				srv.Shutdown(ctx)
			}(server)
		}
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
		log.Println("\x1b[32;1mSuccess shutdown all server & worker\x1b[0m")
	case <-forceShutdown:
		fmt.Println("\x1b[31;1mForce shutdown server & worker\x1b[0m")
		cancel()
	case <-ctx.Done():
		fmt.Println("\x1b[31;1mContext timeout\x1b[0m")
		return
	}
}
