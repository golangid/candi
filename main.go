package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/internal/app"
)

func main() {
	rootApp, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer func() {
		cancel()
		if r := recover(); r != nil {
			fmt.Println("Failed to start service:", r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(ctx, rootApp)
	defer cfg.Exit(ctx)

	service := app.New(cfg)

	// serve http server
	go service.ServeHTTP()

	// serve grpc server
	go service.ServeGRPC()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	service.Shutdown(ctx)
}
