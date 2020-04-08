package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/middleware"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

// App service
type App struct {
	serviceName   constant.Service
	config        *config.Config
	modules       []factory.ModuleFactory
	httpServer    *echo.Echo
	grpcServer    *grpc.Server
	kafkaConsumer sarama.ConsumerGroup
}

// New service app
func New(service factory.ServiceFactory) *App {
	defer log.Printf("Starting %s service\n", service.Name())

	cfg := service.GetConfig()
	mw := middleware.NewMiddleware(cfg)
	params := &base.ModuleParam{
		Config:     cfg,
		Middleware: mw,
	}

	appInstance := new(App)
	appInstance.serviceName = service.Name()
	appInstance.config = cfg
	appInstance.modules = service.Modules(params)

	if config.GlobalEnv.UseHTTP {
		appInstance.httpServer = echo.New()
	}

	if config.GlobalEnv.UseGRPC {
		// init grpc server
		appInstance.grpcServer = grpc.NewServer(
			grpc.MaxSendMsgSize(200*int(helper.MByte)), grpc.MaxRecvMsgSize(200*int(helper.MByte)),
			grpc.UnaryInterceptor(mw.GRPCAuth),
			grpc.StreamInterceptor(mw.GRPCAuthStream),
		)
	}

	if config.GlobalEnv.UseGraphQL {
		appInstance.httpServer.Any("/graphql", echo.WrapHandler(appInstance.graphqlHandler()))
	}

	if config.GlobalEnv.UseKafka {
		// init kafka consumer
		kafkaConsumer, err := sarama.NewConsumerGroup(config.GlobalEnv.Kafka.Brokers, config.GlobalEnv.Kafka.ConsumerGroup, cfg.KafkaConsumerConfig)
		if err != nil {
			log.Panicf("Error creating consumer group client: %v", err)
		}
		appInstance.kafkaConsumer = kafkaConsumer
	}

	return appInstance
}

// Run start app
func (a *App) Run(ctx context.Context) {
	quitSignal := make(chan os.Signal, 1)

	hasServiceHandlerRunning := a.httpServer != nil || a.grpcServer != nil || a.kafkaConsumer != nil
	if !hasServiceHandlerRunning {
		log.Println("No service handler running, shutdown...")
		goto shutdown
	}

	// serve http server
	if a.httpServer != nil {
		go a.ServeHTTP()
	}

	// serve grpc server
	if a.grpcServer != nil {
		go a.ServeGRPC()
	}

	// serve kafka consumer
	if a.kafkaConsumer != nil {
		go a.KafkaConsumer()
	}

	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	<-quitSignal

shutdown:
	a.Shutdown(ctx)
}

// Shutdown graceful shutdown all server, panic if there is still a process running when the request exceed given timeout in context
func (a *App) Shutdown(ctx context.Context) {
	fmt.Println()

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			panic(err)
		}
	}

	if a.grpcServer != nil {
		log.Println("Stopping GRPC server...")
		a.grpcServer.GracefulStop()
	}

	if a.kafkaConsumer != nil {
		log.Println("Stopping kafka consumer...")
		a.kafkaConsumer.Close()
	}
}
