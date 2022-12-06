package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golangid/candi/codebase/factory"
)

// App service
type App struct {
	service factory.ServiceFactory
}

// New service app
func New(service factory.ServiceFactory) *App {

	return &App{
		service: service,
	}
}

// Run start app
func (a *App) Run() {

	if err := a.checkRequired(); err != nil {
		panic(err)
	}

	errServe := make(chan error)
	for _, app := range a.service.GetApplications() {
		go func(srv factory.AppServerFactory) {
			defer func() {
				if r := recover(); r != nil {
					errServe <- fmt.Errorf("%v", r)
				}
			}()
			srv.Serve()
		}(app)
	}

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)

	log.Printf("Application \x1b[32;1m%s\x1b[0m ready to run\n\n", a.service.Name())

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
		for _, server := range a.service.GetApplications() {
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
	}
}

func (a *App) checkRequired() (err error) {

	if len(a.service.GetApplications()) == 0 {
		return errors.New("No server/worker running")
	}

	return
}
