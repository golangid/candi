# Example

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"time"

	taskqueueworker "pkg.agungdp.dev/candi/codebase/app/task_queue_worker"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/logger"
)

// TaskQueueHandler struct
type TaskQueueHandler struct {
}

// NewTaskQueueHandler constructor
func NewTaskQueueHandler() *TaskQueueHandler {
	return &TaskQueueHandler{}
}

// MountHandlers return map topic to handler func
func (h *TaskQueueHandler) MountHandlers(group *types.WorkerHandlerGroup) {

	group.Add("task-one", h.taskOne)
	group.Add("task-two", h.taskTwo)
}

func (h *TaskQueueHandler) taskOne(ctx context.Context, message []byte) error {
	logger.LogRed("task-one: " + string(message))
	return &taskqueueworker.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error",
	}
}

func (h *TaskQueueHandler) taskTwo(ctx context.Context, message []byte) error {
	logger.LogYellow("task-two: " + string(message))
	return &taskqueueworker.ErrorRetrier{
		Delay:   3 * time.Second,
		Message: "Error",
	}
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
			types.TaskQueue: workerhandler.NewTaskQueueHandler(),
		},
	}
}

// ...another method
```

## Add task in each usecase module

```go
package usecase

import (
	"context"
	"log"

	taskqueueworker "pkg.agungdp.dev/candi/codebase/app/task_queue_worker"
)

func someUsecase() {
	// add task queue for `task-one` with 5 retry
	if err := taskqueueworker.AddJob("task-one", 5, `{"params": "test-one"}`); err != nil {
		log.Println(err)
	}

	// add task queue for `task-two` with 5 retry
	if err := taskqueueworker.AddJob("task-two", 5, `{"params": "test-two"}`); err != nil {
		log.Println(err)
	}
}
```