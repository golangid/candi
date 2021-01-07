package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/soheilhy/cmux"
	cronworker "pkg.agungdwiprasetyo.com/candi/codebase/app/cron_worker"
	graphqlserver "pkg.agungdwiprasetyo.com/candi/codebase/app/graphql_server"
	grpcserver "pkg.agungdwiprasetyo.com/candi/codebase/app/grpc_server"
	kafkaworker "pkg.agungdwiprasetyo.com/candi/codebase/app/kafka_worker"
	redisworker "pkg.agungdwiprasetyo.com/candi/codebase/app/redis_worker"
	restserver "pkg.agungdwiprasetyo.com/candi/codebase/app/rest_server"
	taskqueueworker "pkg.agungdwiprasetyo.com/candi/codebase/app/task_queue_worker"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/config/env"
)

// App service
type App struct {
	servers []factory.AppServerFactory
	cMux    cmux.CMux
}

// New service app
func New(service factory.ServiceFactory) *App {
	log.Printf("Starting \x1b[32;1m%s\x1b[0m service\n\n", service.Name())

	appInstance := new(App)

	if env.BaseEnv().UseKafkaConsumer {
		appInstance.servers = append(appInstance.servers, kafkaworker.NewWorker(service))
	}

	if env.BaseEnv().UseCronScheduler {
		appInstance.servers = append(appInstance.servers, cronworker.NewWorker(service))
	}

	if env.BaseEnv().UseTaskQueueWorker {
		appInstance.servers = append(appInstance.servers, taskqueueworker.NewWorker(service))
	}

	if env.BaseEnv().UseRedisSubscriber {
		appInstance.servers = append(appInstance.servers, redisworker.NewWorker(service))
	}

	useSharedListener := (env.BaseEnv().UseREST || env.BaseEnv().UseGraphQL) && (env.BaseEnv().UseGRPC && env.BaseEnv().GRPCPort == 0)
	if useSharedListener {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", env.BaseEnv().HTTPPort))
		if err != nil {
			panic(err)
		}
		appInstance.cMux = cmux.New(listener)
		appInstance.cMux.SetReadTimeout(30 * time.Second)
	}

	if env.BaseEnv().UseREST {
		appInstance.servers = append(appInstance.servers, restserver.NewServer(service, appInstance.cMux))
	}

	if env.BaseEnv().UseGRPC {
		appInstance.servers = append(appInstance.servers, grpcserver.NewServer(service, appInstance.cMux))
	}

	if !env.BaseEnv().UseREST && env.BaseEnv().UseGraphQL {
		appInstance.servers = append(appInstance.servers, graphqlserver.NewServer(service, appInstance.cMux))
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

	if a.cMux != nil {
		go func() { a.cMux.Serve() }()
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
		for _, server := range a.servers {
			server.Shutdown(ctx)
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
		log.Println("\x1b[32;1mSuccess shutdown all server & worker\x1b[0m")
	case <-forceShutdown:
		log.Println("\x1b[31;1mForce shutdown server & worker\x1b[0m")
		cancel()
	case <-ctx.Done():
		log.Println("\x1b[31;1mContext timeout\x1b[0m")
		return
	}
}
