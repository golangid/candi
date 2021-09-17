# Broker

Include default broker (Kafka & RabbitMQ), or other broker (GCP PubSub, STOMP/AMQ) can be found in [candi plugin](https://github.com/agungdwiprasetyo/candi-plugin).

## Kafka

**Register Kafka broker in service config**

Modify `configs/configs.go` in your service

```go
package configs

import (
	"github.com/golangid/candi/broker"
...

// LoadServiceConfigs load selected dependency configuration in this service
func LoadServiceConfigs(baseCfg *config.Config) (deps dependency.Dependency) {
	
		...

		brokerDeps := broker.InitBrokers(
			broker.NewKafkaBroker(),
		)

		... 
}
```

If you want to use Kafka consumer, just set `USE_KAFKA_CONSUMER=true` in environment variable, and follow [this example](https://github.com/agungdwiprasetyo/candi/tree/master/codebase/app/kafka_worker).

If you want to use Kafka publisher in your usecase, follow this example code:

```go
package usecase

import (
	"context"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

type usecaseImpl {
	kafkaPub interfaces.Publisher
}

func NewUsecase(deps dependency.Dependency) Usecase {
	return &usecaseImpl{
		kafkaPub: deps.GetBroker(types.Kafka).GetPublisher(),
	}
}

func (uc *usecaseImpl) UsecaseToPublishMessage(ctx context.Context) error {
	err := uc.kafkaPub.PublishMessage(ctx, &candishared.PublisherArgument{
		Topic:  "example-topic",
		Data:   "hello world",
	})
	return err
}
```

## RabbitMQ

**Register RabbitMQ broker in service config**

Modify `configs/configs.go` in your service

```go
package configs

import (
	"github.com/golangid/candi/broker"
...

// LoadServiceConfigs load selected dependency configuration in this service
func LoadServiceConfigs(baseCfg *config.Config) (deps dependency.Dependency) {
	
		...

		brokerDeps := broker.InitBrokers(
			broker.NewRabbitMQBroker(),
		)

		... 
}
```

If you want to use RabbitMQ consumer, just set `USE_RABBITMQ_CONSUMER=true` in environment variable, and follow [this example](https://github.com/agungdwiprasetyo/candi/tree/master/codebase/app/rabbitmq_worker).

If you want to use RabbitMQ publisher in your usecase, follow this example code:

```go
package usecase

import (
	"context"

	"github.com/golangid/candi/broker"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

type usecaseImpl {
	rabbitmqPub interfaces.Publisher
}

func NewUsecase(deps dependency.Dependency) Usecase {
	return &usecaseImpl{
		rabbitmqPub: deps.GetBroker(types.RabbitMQ).GetPublisher(),
	}
}

func (uc *usecaseImpl) UsecaseToPublishMessage(ctx context.Context) error {
	err := uc.rabbitmqPub.PublishMessage(ctx, &candishared.PublisherArgument{
		Topic:  "example-topic",
		Data:   "hello world"
		Header: map[string]interface{}{
			broker.RabbitMQDelayHeader: 5000, // if you want set delay consume your message by active consumer for 5 seconds
		},
	})
	return err
}
```
