# Example

## Create delivery handler

```go
package workerhandler

import (
	"time"

	"github.com/golangid/candi/candishared"
	taskqueueworker "github.com/golangid/candi/codebase/app/task_queue_worker"
	"github.com/golangid/candi/codebase/factory/types"
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

func (h *TaskQueueHandler) taskOne(eventContext *candishared.EventContext) error {

	fmt.Printf("executing task '%s' has been %s retry, with message: %s\n",
		eventContext.HandlerRoute(),
		eventContext.Header()["retries"],
		eventContext.Message(),
	)

	return &taskqueueworker.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error",
	}
}

func (h *TaskQueueHandler) taskTwo(eventContext *candishared.EventContext) error {

	fmt.Printf("executing task '%s' has been %s retry, with message: %s\n",
		eventContext.HandlerRoute(),
		eventContext.Header()["retries"],
		eventContext.Message(),
	)

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
			types.TaskQueue: workerhandler.NewTaskQueueHandler(),
		},
	}
}

// ...another method
```

## Add task in each usecase module

### From internal service (same runtime)

```go
package usecase

import (
	"context"
	"log"

	taskqueueworker "github.com/golangid/candi/codebase/app/task_queue_worker"
)

func someUsecase(ctx context.Context) {
	// add task queue for `{{task_name}}` with 5 retry
	if err := taskqueueworker.AddJob(ctx, "{{task-queue-worker-host}}", &taskqueueworker.AddJobRequest{
	TaskName: "{{task_name}}", MaxRetry: 5, Args: []byte(`{{arguments/message}}`),
	}); err != nil {
		log.Println(err)
	}
}
```

### Or if running on a separate server

- Via GraphQL API:

`POST {{task-queue-worker-host}}/graphql`
```
mutation addJob {
  add_job(
	  param: {
		task_name: "{{task_name}}"
		max_retry: 5
		args: "{\"params\": \"test-one\"}"
	  }
  )
}
```

- cURL:
```
curl --location --request POST '{{task-queue-worker-host}}/graphql' \
--header 'Content-Type: application/json' \
--data-raw '{
    "operationName": "addJob",
    "variables": {
        "param": {
            "task_name": "{{task_name}}",
            "max_retry": 1,
            "args": "{{arguments/message}}"
        }
    },
    "query": "mutation addJob($param: AddJobInputResolver!) {\n  add_job(param: $param)\n}\n"
}'
```

- Direct call function:
```go
// add task queue for `task-one` via HTTP request
jobID, err := taskqueueworker.AddJobViaHTTPRequest(context.Background(), "{{task-queue-worker-host}}", &taskqueueworker.AddJobRequest{
	TaskName: "{{task_name}}", MaxRetry: 1, Args: []byte(`{{arguments/message}}`),
})
if err != nil {
	log.Println(err)
}
fmt.Println("Queued job id is ", jobID)
```

- Auto call to external worker via HTTP request using basic function `AddJob()`, please set configuration:
```go
taskqueueworker.SetExternalWorkerHost({{task-queue-worker-host}})
```

```go
jobID, err := taskqueueworker.AddJob(context.Background(), &taskqueueworker.AddJobRequest{
	TaskName: "{{task_name}}", MaxRetry: 1, Args: []byte(`{{arguments/message}}`),
})
fmt.Println("Queued job id is ", jobID)
```
