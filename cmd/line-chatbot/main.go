package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	linechatbot "agungdwiprasetyo.com/backend-microservices/internal/line-chatbot"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/app"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
)

const (
	serviceName = "line-chatbot"
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

	service := linechatbot.NewService(serviceName, base.InitDependency(cfg))
	app.New(service).Run(ctx)
}
