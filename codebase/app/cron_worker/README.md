# Example

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"fmt"
	"time"

	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/candihelper"
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

	group.Add(candihelper.CronJobKeyToString("push-notif", "message", "30s"), h.handlePushNotif)
	group.Add(candihelper.CronJobKeyToString("heavy-push-notif", "message", "22:43:07"), h.handleHeavyPush)
}

func (h *CronHandler) handlePushNotif(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "CronDelivery-HandlePushNotif")
	defer trace.Finish()

	logger.LogI("processing")
	logger.LogI("done")
	return nil
}

func (h *CronHandler) handleHeavyPush(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "CronDelivery-HandleHeavyPush")
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
