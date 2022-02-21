# Example

This is example for create Postgres Event Listener (CDC), inspired by [**Hasura Event Triggers**](https://hasura.io/docs/latest/graphql/core/event-triggers/index.html)

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

// PostgresListenerHandler struct
type PostgresListenerHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewPostgresListenerHandler constructor
func NewPostgresListenerHandler(uc usecase.Usecase, validator interfaces.Validator) *PostgresListenerHandler {
	return &PostgresListenerHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *PostgresListenerHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("table-names", h.handleDataChange) // listen change data capture on table "table-names"
}

func (h *PostgresListenerHandler) handleDataChange(eventContext *candishared.EventContext) error {
	trace := tracer.StartTrace(eventContext.Context(), "DeliveryPostgresListener:HandleDataChange")
	defer trace.Finish()

	fmt.Printf("data change on table 'table-names' detected: %s\n", eventContext.Message())
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
			types.PostgresListener: workerhandler.NewPostgresListenerHandler(usecaseUOW.User(), deps.GetValidator()),
		},
	}
}

// ...another method
```

## Message in JSON Payload

```
{
	"event_id": "<md5-hash-unique-event-id>",
	"table": "<table-name>",
	"action": "<operation-name>", // INSERT, UPDATE, or DELETE
	"data": {
		"old": <old-column-values-object>,
		"new": <new-column-values-object>
	}
}
```
