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

// Option app option
type Option func(*App)

// SetShutdownTimeout set timeout for graceful shutdown
func SetShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(a *App) {
		a.shutdownTimeout = shutdownTimeout
	}
}

// SetAfterShutdown set do after shutdown
func SetAfterShutdown(do func()) Option {
	return func(a *App) {
		a.afterShutdown = do
	}
}

// SetQuitSignalTrigger option
func SetQuitSignalTrigger(quitSignalTriggers []os.Signal) Option {
	return func(a *App) {
		a.quitSignalTriggers = quitSignalTriggers
	}
}

// App service
type App struct {
	done               chan struct{}
	afterShutdown      func()
	shutdownTimeout    time.Duration
	quitSignal         chan os.Signal
	quitSignalTriggers []os.Signal
	service            factory.ServiceFactory
}

// New init new service app
func New(service factory.ServiceFactory, opts ...Option) *App {
	app := &App{
		service:            service,
		done:               make(chan struct{}, 1),
		shutdownTimeout:    1 * time.Minute,
		quitSignal:         make(chan os.Signal, 1),
		quitSignalTriggers: []os.Signal{os.Interrupt, syscall.SIGTERM},
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Run start app
func (a *App) Run() {
	if err := a.checkRequired(); err != nil {
		log.Panic(err)
	}

	errServe := make(chan error)
	checkExist := make(map[string]struct{})
	for _, app := range a.service.GetApplications() {
		if _, ok := checkExist[app.Name()]; ok {
			log.Panicf("Register application: %s has been registered", app.Name())
		}
		checkExist[app.Name()] = struct{}{}
		go func(srv factory.AppServerFactory) {
			defer func() {
				if r := recover(); r != nil {
					errServe <- fmt.Errorf("%v", r)
				}
			}()
			srv.Serve()
		}(app)
	}

	signal.Notify(a.quitSignal, a.quitSignalTriggers...)

	log.Printf("Application \x1b[32;1m%s\x1b[0m ready to run\n\n", a.service.Name())

	select {
	case e := <-errServe:
		log.Panic(e)
	case <-a.quitSignal:
		a.shutdown()
	}
}

// Shutdown for manual trigger for shutdown
func (a *App) Shutdown() {
	a.quitSignal <- os.Interrupt
	<-a.done
}

// shutdown graceful shutdown all server, panic if there is still a process running when the request exceed given timeout in context
func (a *App) shutdown() {
	fmt.Println("\x1b[34;1mGracefully shutdown... (press Ctrl+C again to force)\x1b[0m")

	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer func() {
		cancel()
		if a.afterShutdown != nil {
			a.afterShutdown()
		}
		a.done <- struct{}{}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, svc := range a.service.GetApplications() {
			svc.Shutdown(ctx)
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
		log.Println("\x1b[32;1mSuccess shutdown all application\x1b[0m")
	case <-a.quitSignal:
		log.Println("\x1b[31;1mForce shutdown application\x1b[0m")
		os.Exit(0)
	case <-ctx.Done():
		log.Printf("\x1b[31;1mShutdown timeout after %s. Force shutdown\x1b[0m", a.shutdownTimeout.String())
		os.Exit(0)
	}
}

func (a *App) checkRequired() (err error) {
	if len(a.service.GetApplications()) == 0 {
		return errors.New("No server/worker running")
	}
	return
}
