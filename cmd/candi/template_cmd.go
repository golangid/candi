package main

const cmdMainTemplate = `package main

import (
	"fmt"
	"runtime/debug"

	"pkg.agungdwiprasetyo.com/candi/codebase/app"
	"pkg.agungdwiprasetyo.com/candi/config"

	service "{{.ServiceName}}/internal"
)

const serviceName = "{{.ServiceName}}"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\x1b[31;1mFailed to start %s service: %v\x1b[0m\n", serviceName, r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(fmt.Sprintf("cmd/%s/", serviceName))
	defer cfg.Exit()

	srv := service.NewService(serviceName, cfg)
	app.New(srv).Run()
}

`
