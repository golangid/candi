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
func (h *KafkaHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		"{{$.module}}": h.handleTest,
	}
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
func (h *CronHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		helper.CronJobKeyToString("sample", "10s"): h.handleSample,
	}
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
func (h *RedisHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		"{{$.module}}-sample": h.handleSample,
	}
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
func (h *TaskQueueHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		"{{$.module}}-task-one": h.taskOne,
		"{{$.module}}-task-two": h.taskTwo,
	}
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
