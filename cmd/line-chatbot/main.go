package main

import (
	"fmt"
	"runtime/debug"

	"agungdwiprasetyo.com/backend-microservices/config"
	linechatbot "agungdwiprasetyo.com/backend-microservices/internal/line-chatbot"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/app"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
)

const (
	serviceName = "line-chatbot"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Failed to start %s service: %v\n", serviceName, r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(fmt.Sprintf("cmd/%s/", serviceName))
	defer cfg.Exit()

	service := linechatbot.NewService(serviceName, base.InitDependency(cfg))
	app.New(service).Run()
}
