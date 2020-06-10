package main

const cmdMainTemplate = `package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"{{.PackageName}}/config"
	"{{.PackageName}}/internal/app"
	"{{.PackageName}}/internal/factory/base"
	service "{{.PackageName}}/internal/services/{{.ServiceName}}"
)

const (
	serviceName = "{{.ServiceName}}"
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

	srv := service.NewService(serviceName, base.InitDependency(cfg))
	app.New(srv).Run(ctx)
}

`
