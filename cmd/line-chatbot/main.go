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

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/app"
	linechatbot "agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot"
	"github.com/joho/godotenv"
)

const (
	appLocation = "cmd/line-chatbot"
)

func main() {
	rootApp, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	rootApp = strings.TrimSuffix(rootApp, appLocation) // trim this path location, for cleaning root path
	// load additional env
	if err := godotenv.Load(appLocation + "/.env"); err != nil {
		log.Println("additional env not declared")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer func() {
		cancel()
		if r := recover(); r != nil {
			fmt.Println("Failed to start linechatbot service:", r)
			fmt.Printf("Stack trace: \n%s\n", debug.Stack())
		}
	}()

	cfg := config.Init(ctx, rootApp)
	defer cfg.Exit(ctx)

	service := linechatbot.NewService(cfg)
	app.New(service).Run(ctx)
}
