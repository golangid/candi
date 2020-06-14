package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/app"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	service "agungdwiprasetyo.com/backend-microservices/internal/services/user-service"
)

const (
	serviceName = "user-service"
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

	cfg := config.Init(ctx, fmt.Sprintf("cmd/%s/", serviceName))
	defer cfg.Exit(ctx)

	srv := service.NewService(serviceName, base.InitDependency(cfg))
	app.New(srv).Run(ctx)
}

