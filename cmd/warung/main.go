package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/internal/app"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/warung"
)

func main() {
	rootApp, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	rootApp = strings.TrimSuffix(rootApp, "/cmd/warung") // trim this path location, for cleaning root path

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer func() {
		cancel()
		if r := recover(); r != nil {
			fmt.Println("Failed to start warung service:", r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(ctx, rootApp)
	defer cfg.Exit(ctx)

	service := warung.NewService(cfg)
	app.New(service).Run(ctx)
}
