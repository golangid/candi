package main

const (
	deliveryKafkaTemplate = `package workerhandler

import (
	"context"
	"fmt"

	"{{.PackageName}}/pkg/codebase/factory/types"
)

// KafkaHandler struct
type KafkaHandler struct {
}

// NewKafkaHandler constructor
func NewKafkaHandler() *KafkaHandler {
	return &KafkaHandler{}
}

// MountHandlers return map topic to handler func
func (h *KafkaHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{$.module}}", h.handleTest)
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) handleTest(ctx context.Context, message []byte) error {
	fmt.Printf("message consumed by module {{$.module}}. message: %s\n", string(message))
	return nil
}
`

	deliveryCronTemplate = `package workerhandler

import (
	"context"
	"fmt"

	"{{.PackageName}}/pkg/codebase/factory/types"
	"{{.PackageName}}/pkg/helper"
)

// CronHandler struct
type CronHandler struct {
}

// NewCronHandler constructor
func NewCronHandler() *CronHandler {
	return &CronHandler{}
}

// MountHandlers return map topic to handler func
func (h *CronHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add(helper.CronJobKeyToString("sample", "10s"), h.handleSample)
}

func (h *CronHandler) handleSample(ctx context.Context, message []byte) error {
	fmt.Println("cron: execute sample")
	return nil
}
`

	deliveryRedisTemplate = `package workerhandler

import (
	"context"
	"fmt"

	"{{.PackageName}}/pkg/codebase/factory/types"
)

// RedisHandler struct
type RedisHandler struct {
}

// NewRedisHandler constructor
func NewRedisHandler() *RedisHandler {
	return &RedisHandler{
	}
}

// MountHandlers return map topic to handler func
func (h *RedisHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{$.module}}-sample", h.handleSample)
}

func (h *RedisHandler) handleSample(ctx context.Context, message []byte) error {
	fmt.Println("redis subs: execute sample")
	return nil
}
`

	deliveryTaskQueueTemplate = `package workerhandler

import (
	"context"
	"time"

	taskqueueworker "{{.PackageName}}/pkg/codebase/app/task_queue_worker"
	"{{.PackageName}}/pkg/codebase/factory/types"
)

// TaskQueueHandler struct
type TaskQueueHandler struct {
}

// NewTaskQueueHandler constructor
func NewTaskQueueHandler() *TaskQueueHandler {
	return &TaskQueueHandler{
	}
}

// MountHandlers return map topic to handler func
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
