package main

import (
	"fmt"
	"runtime/debug"

	"agungdwiprasetyo.com/backend-microservices/config"
	linechatbot "agungdwiprasetyo.com/backend-microservices/internal/line-chatbot"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/app"
)

const (
	serviceName = "line-chatbot"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\x1b[31;1mFailed to start %s service: %v\x1b[0m\n", serviceName, r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(fmt.Sprintf("cmd/%s/", serviceName))
	defer cfg.Exit()

	service := linechatbot.NewService(serviceName, cfg)
	app.New(service).Run()
}
