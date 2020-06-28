package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

const timeout int64 = 10 // in seconds

var env Env

// Config app
type Config struct {
	closers []interfaces.Closer
}

// Init app config
func Init(rootApp string) *Config {
	loadBaseEnv(rootApp, &env)
	return &Config{}
}

// BaseEnv get global basic environment
func BaseEnv() Env {
	return env
}

// LoadFunc load selected dependency with context timeout
func (c *Config) LoadFunc(depsFunc func(context.Context) []interfaces.Closer) {
	// set timeout for init configuration
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	result := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("Failed init configuration :=> %v", r)
			}
			close(result)
		}()

		c.closers = depsFunc(ctx)
	}()

	// with timeout to init configuration
	select {
	case <-ctx.Done():
		panic(fmt.Errorf("Timeout to load selected dependencies: %v", ctx.Err()))
	case err := <-result:
		if err != nil {
			panic(err)
		}
		return
	}
}

// Exit close all connection
func (c *Config) Exit() {
	// set timeout for close all configuration
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	fmt.Println()

	result := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("Failed close connection :=> %v", r)
			}
			close(result)
		}()

		for _, cl := range c.closers {
			cl.Disconnect(ctx)
		}
	}()

	// with timeout to close all configuration
	select {
	case <-ctx.Done():
		panic(fmt.Errorf("Timeout to close all selected dependencies connection: %v", ctx.Err()))
	case err := <-result:
		if err != nil {
			panic(err)
		}
		log.Println("\x1b[32;1mSuccess close all config connection\x1b[0m")
		return
	}
}
