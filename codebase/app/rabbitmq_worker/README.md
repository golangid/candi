# Example

This is example for create RabbitMQ consumer handler in delivery layer.

RabbitMQ Consumer with delayed message feature, please add this [**plugin**](https://github.com/agungdwiprasetyo/docker-apps/tree/master/rabbitmq/plugins) to RabbitMQ broker.

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"encoding/json"
	
	"example.service/internal/modules/examplemodule/delivery/workerhandler"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/tracer"
)

// RabbitMQHandler struct
type RabbitMQHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewRabbitMQHandler constructor
func NewRabbitMQHandler(uc usecase.Usecase, validator interfaces.Validator) *RabbitMQHandler {
	return &RabbitMQHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *RabbitMQHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("example-queue", h.handleQueue) // consume queue "example-queue"
}

func (h *RabbitMQHandler) handleQueue(eventContext *candishared.EventContext) error {
	trace := tracer.StartTrace(eventContext.Context(), "DeliveryRabbitMQ:HandleQueue")
	defer trace.Finish()

	log.Printf("message consumed. message: %s\n", eventContext.Message())
	// call usecase
	return nil
}

```

## Register in module

```go
package examplemodule

import (
	"example.service/internal/modules/examplemodule/delivery/workerhandler"

	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

type Module struct {
	// ...another delivery handler
	workerHandlers map[types.Worker]interfaces.WorkerHandler
}

func NewModules(deps dependency.Dependency) *Module {
	return &Module{
		workerHandlers: map[types.Worker]interfaces.WorkerHandler{
			// ...another worker handler
			// ...
			types.RabbitMQ: workerhandler.NewRabbitMQHandler(usecaseUOW.User(), deps.GetValidator()),
		},
	}
}

// ...another method
```
