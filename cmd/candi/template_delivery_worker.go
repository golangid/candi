package main

const (
	deliveryKafkaTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackageName}}/codebase/factory/types"
)

// KafkaHandler struct
type KafkaHandler struct {
}

// NewKafkaHandler constructor
func NewKafkaHandler() *KafkaHandler {
	return &KafkaHandler{}
}

// MountHandlers mount handler group
func (h *KafkaHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{$.module}}", h.handle{{clean (upper $.module)}}) // handling topic "{{$.module}}"
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) handle{{clean (upper $.module)}}(ctx context.Context, message []byte) error {
	fmt.Printf("message consumed by module {{$.module}}. message: %s\n", string(message))
	return nil
}
`

	deliveryCronTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackageName}}/candihelper"
	"{{.PackageName}}/codebase/factory/types"
)

// CronHandler struct
type CronHandler struct {
}

// NewCronHandler constructor
func NewCronHandler() *CronHandler {
	return &CronHandler{}
}

// MountHandlers mount handler group
func (h *CronHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add(candihelper.CronJobKeyToString("{{$.module}}-scheduler", "10s"), h.handle{{clean (upper $.module)}})
}

func (h *CronHandler) handle{{clean (upper $.module)}}(ctx context.Context, message []byte) error {
	fmt.Println("cron: execute in module {{$.module}}")
	return nil
}
`

	deliveryRedisTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackageName}}/codebase/factory/types"
)

// RedisHandler struct
type RedisHandler struct {
}

// NewRedisHandler constructor
func NewRedisHandler() *RedisHandler {
	return &RedisHandler{
	}
}

// MountHandlers mount handler group
func (h *RedisHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{$.module}}-sample", h.handle{{clean (upper $.module)}})
}

func (h *RedisHandler) handle{{clean (upper $.module)}}(ctx context.Context, message []byte) error {
	fmt.Println("redis subs: execute sample")
	return nil
}
`

	deliveryTaskQueueTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"time"

	taskqueueworker "{{.PackageName}}/codebase/app/task_queue_worker"
	"{{.PackageName}}/codebase/factory/types"
)

// TaskQueueHandler struct
type TaskQueueHandler struct {
}

// NewTaskQueueHandler constructor
func NewTaskQueueHandler() *TaskQueueHandler {
	return &TaskQueueHandler{
	}
}

// MountHandlers mount handler group
func (h *TaskQueueHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{$.module}}-task-one", h.taskOne)
	group.Add("{{$.module}}-task-two", h.taskTwo)
}

func (h *TaskQueueHandler) taskOne(ctx context.Context, message []byte) error {
	return &taskqueueworker.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error one",
	}
}

func (h *TaskQueueHandler) taskTwo(ctx context.Context, message []byte) error {
	return &taskqueueworker.ErrorRetrier{
		Delay:   3 * time.Second,
		Message: "Error two",
	}
}
`
)
