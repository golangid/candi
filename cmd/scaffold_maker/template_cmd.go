package main

const cmdMainTemplate = `package main

import (
	"fmt"
	"runtime/debug"

	"{{.PackageName}}/config"
	service "{{.PackageName}}/internal/{{.ServiceName}}"
	"{{.PackageName}}/pkg/codebase/app"
	"{{.PackageName}}/pkg/codebase/factory/base"
)

const (
	serviceName = "{{.ServiceName}}"
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

	srv := service.NewService(serviceName, base.InitDependency(cfg))
	app.New(srv).Run()
}

`
