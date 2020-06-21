package main

import (
	"fmt"
	"runtime/debug"

	"agungdwiprasetyo.com/backend-microservices/config"
	service "agungdwiprasetyo.com/backend-microservices/internal/user-service"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/app"
)

const (
	serviceName = "user-service"
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

	srv := service.NewService(serviceName, cfg)
	app.New(srv).Run()
}
