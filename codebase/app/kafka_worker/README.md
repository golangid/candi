# Example

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"encoding/json"
	
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/tracer"
)

// KafkaHandler struct
type KafkaHandler struct {
}

// NewKafkaHandler constructor
func NewKafkaHandler() *KafkaHandler {
	return &KafkaHandler{}
}

// MountHandlers return group map topic to handler func
func (h *KafkaHandler) MountHandlers(group *types.WorkerHandlerGroup) {

	group.Add("example-topic", h.handleExampleTopic) // handling consume topic "example-topic"
}

func (h *KafkaHandler) handleExampleTopic(eventContext *candishared.EventContext) error {
	trace := tracer.StartTrace(eventContext.Context(), "DeliveryKafkaConsumer:HandleExampleTopic")
	defer trace.Finish()

	// process usecase
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
			types.Kafka: workerhandler.NewKafkaHandler(),
		},
	}
}

// ...another method
```
