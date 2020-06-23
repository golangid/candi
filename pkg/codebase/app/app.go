package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	graphqlserver "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/graphql_server"
	grpcserver "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/grpc_server"
	kafkaworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/kafka_worker"
	restserver "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/rest_server"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	_ "agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

// App service
type App struct {
	servers []factory.AppServerFactory
}

// New service app
func New(service factory.ServiceFactory) *App {
	log.Printf("Starting \x1b[34;1m%s\x1b[0m service\n", service.Name())

	appInstance := new(App)
	if config.BaseEnv().UseREST {
		appInstance.servers = append(appInstance.servers, restserver.NewServer(service))
	}

	if config.BaseEnv().UseGRPC {
		appInstance.servers = append(appInstance.servers, grpcserver.NewServer(service))
	}

	if config.BaseEnv().UseGraphQL {
		appInstance.servers = append(appInstance.servers, graphqlserver.NewServer(service))
	}

	if config.BaseEnv().UseKafkaConsumer {
		appInstance.servers = append(appInstance.servers, kafkaworker.NewWorker(service))
	}

	return appInstance
}

// Run start app
func (a *App) Run() {
	if len(a.servers) == 0 {
		panic("No server handler running")
	}

	for _, server := range a.servers {
		go server.Serve()
	}

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	<-quitSignal

	a.shutdown()
}

// graceful shutdown all server, panic if there is still a process running when the request exceed given timeout in context
func (a *App) shutdown() {
	println("Graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for _, server := range a.servers {
		server.Shutdown(ctx)
	}

	log.Println("Success shutdown all server")
}
