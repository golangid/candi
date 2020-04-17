package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/app"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung"
)

const (
	serviceName = "warung"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer func() {
		cancel()
		if r := recover(); r != nil {
			fmt.Printf("Failed to start %s service: %v\n", serviceName, r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(ctx, "cmd/"+serviceName)
	defer cfg.Exit(ctx)

	service := warung.NewService(cfg, serviceName)
	app.New(service).Run(ctx)
}
