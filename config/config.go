package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"github.com/Shopify/sarama"
)

var env Env

// Config app
type Config struct {
	KafkaConfig *sarama.Config

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

// Load load selected dependency with context timeout
func (c *Config) Load(selectedDepsFunc ...func(ctx context.Context) interfaces.Closer) {
	// set timeout for init configuration
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	result := make(chan error)
	go func(loaderFuncs ...func(ctx context.Context) interfaces.Closer) {
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("Failed init configuration :=> %v", r)
			}
			close(result)
		}()

		for _, fn := range loaderFuncs {
			c.closers = append(c.closers, fn(ctx))
		}
	}(selectedDepsFunc...)

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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for _, cl := range c.closers {
		if cl != nil {
			cl.Disconnect(ctx)
		}
	}

	log.Println("\x1b[32;1mConfig: Success close all connection\x1b[0m")
}
