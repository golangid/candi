package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/internal/app"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding"
)

func main() {
	rootApp, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	rootApp = strings.TrimSuffix(rootApp, "/cmd/wedding") // trim this path location, for cleaning root path

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer func() {
		cancel()
		if r := recover(); r != nil {
			fmt.Println("Failed to start wedding service:", r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(ctx, rootApp)
	defer cfg.Exit(ctx)

	service := wedding.NewService(cfg)
	app := app.New(service)

	// serve http server
	go app.ServeHTTP()

	// serve grpc server
	go app.ServeGRPC()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	app.Shutdown(ctx)
}
