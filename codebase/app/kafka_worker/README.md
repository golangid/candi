# Example

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"encoding/json"
	
	"pkg.agungdwiprasetyo.com/gendon/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/gendon/tracer"
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

	group.Add("push-notif", h.handlePushNotif)
}

func (h *KafkaHandler) handlePushNotif(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "KafkaDelivery-HandlePushNotif")
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

	"pkg.agungdwiprasetyo.com/gendon/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/gendon/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/gendon/codebase/interfaces"
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
