# Example

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"fmt"
	"time"

	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
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

	group.Add(candihelper.CronJobKeyToString("push-notif", "30s"), h.handlePushNotif)
	group.Add(candihelper.CronJobKeyToString("heavy-push-notif", "22:43:07"), h.handleHeavyPush)
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

	"pkg.agungdp.dev/candi/codebase/factory/dependency"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/codebase/interfaces"
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
