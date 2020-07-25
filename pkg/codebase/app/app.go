package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	cronworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/cron_worker"
	graphqlserver "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/graphql_server"
	grpcserver "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/grpc_server"
	kafkaworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/kafka_worker"
	redisworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/redis_worker"
	restserver "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/rest_server"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
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

	return appInstance
}

// Run start app
func (a *App) Run() {

	if len(a.servers) == 0 {
		panic("No server/worker running")
	}

	for _, server := range a.servers {
		go server.Serve()
	}

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	<-quitSignal

	a.shutdown(quitSignal)
}

// graceful shutdown all server, panic if there is still a process running when the request exceed given timeout in context
func (a *App) shutdown(forceShutdown chan os.Signal) {
	fmt.Println("\x1b[34;1mGracefully shutdown... (press Ctrl+C again to force)\x1b[0m")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, server := range a.servers {
			server.Shutdown(ctx)
		}
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
