# Example

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"fmt"
	"time"

	"github.com/golangid/candi/candishared"
	cronworker "github.com/golangid/candi/codebase/app/cron_worker"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

// CronHandler struct
type CronHandler struct {
}

// NewCronHandler constructor
func NewCronHandler() *CronHandler {
	return &CronHandler{}
}

// MountHandlers return group map topic key to handler func
func (h *CronHandler) MountHandlers(group *types.WorkerHandlerGroup) {

	group.Add(cronworker.CreateCronJobKey("push-notif", "message", "30s"), h.handleJob1)
	group.Add(cronworker.CreateCronJobKey("heavy-push-notif", "message", "22:43:07"), h.handleJob2)
}

func (h *CronHandler) handleJob1(eventContext *candishared.EventContext) error {
	trace := tracer.StartTrace(eventContext.Context(), "DeliveryCronWorker:HandleJob1")
	defer trace.Finish()

	logger.LogI("running...")
	return nil
}

func (h *CronHandler) handleJob2(eventContext *candishared.EventContext) error {
	trace := tracer.StartTrace(eventContext.Context(), "DeliveryCronWorker:HandleJob2")
	defer trace.Finish()

	fmt.Println("processing")
	time.Sleep(30 * time.Second)
	fmt.Println("done")
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
			types.Scheduler: workerhandler.NewCronHandler(),
		},
	}
}

// ...another method
```
