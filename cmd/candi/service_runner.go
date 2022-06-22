package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

func serviceRunner(serviceName string) {
	logger := log.New(os.Stdout, "\x1b[33;1m[service runner] >> \x1b[0m", log.Ldate|log.Ltime)

	var commandList []string
	if serviceName != "" {
		services := strings.Split(serviceName, ",")
		for _, s := range services {
			if _, err := os.Stat("services/" + s); os.IsNotExist(err) {
				log.Fatalf(RedFormat, `ERROR: service "`+s+`" is not exist`)
			}
		}
		commandList = append(commandList, services...)
	} else {
		logger.Printf("\x1b[34;1mRUNNING ALL SERVICE in \"services/\" directory\x1b[0m\n")
		files, err := os.ReadDir("services")
		if err != nil {
			logger.Fatalf(RedFormat, "ERROR: "+err.Error()+" (must in monorepo root)")
		}
		for _, f := range files {
			if f.IsDir() {
				commandList = append(commandList, f.Name())
			}
		}
		if len(commandList) == 0 {
			logger.Fatal("\x1b[31;1mERROR: no service available in \"services/\" directory\x1b[0m\n")
		}
	}

	var cmds []*exec.Cmd
	defer func() {
		for _, c := range cmds {
			c.Wait()
			c.Process.Kill()
		}
	}()

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	for _, serviceName := range commandList {
		logger.Printf("running service \x1b[32;1m%s...\x1b[0m\n", serviceName)
		cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("WORKDIR=services/%s/ go run services/%s/*.go", serviceName, serviceName))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		cmds = append(cmds, cmd)

		if err := cmd.Start(); err != nil {
			logger.Println("ERROR", err)
		}
	}

	<-quitSignal
}
