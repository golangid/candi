package main

const (
	deliveryKafkaTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// KafkaHandler struct
type KafkaHandler struct {
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewKafkaHandler constructor
func NewKafkaHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *KafkaHandler {
	return &KafkaHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *KafkaHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}", h.handle{{clean (upper .ModuleName)}}) // handling topic "{{.ModuleName}}"
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) handle{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryKafka:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	key := candishared.GetValueFromContext(ctx, candishared.ContextKeyWorkerKey).([]byte)
	fmt.Printf("message consumed by module {{.ModuleName}}. key: %s, message: %s\n", key, message)
	return nil
}
`

	deliveryCronTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// CronHandler struct
type CronHandler struct {
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewCronHandler constructor
func NewCronHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *CronHandler {
	return &CronHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *CronHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add(candihelper.CronJobKeyToString("{{.ModuleName}}-scheduler", "message", "10s"), h.handle{{clean (upper .ModuleName)}})
}

func (h *CronHandler) handle{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryCron:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	fmt.Println("cron: execute in module {{.ModuleName}}, message:", string(message))
	return nil
}
`

	deliveryRedisTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// RedisHandler struct
type RedisHandler struct {
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewRedisHandler constructor
func NewRedisHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *RedisHandler {
	return &RedisHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *RedisHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}-sample", h.handle{{clean (upper .ModuleName)}})
}

func (h *RedisHandler) handle{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryRedis:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	fmt.Println("redis subs: execute sample")
	return nil
}
`

	deliveryTaskQueueTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"
	"time"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// TaskQueueHandler struct
type TaskQueueHandler struct {
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewTaskQueueHandler constructor
func NewTaskQueueHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *TaskQueueHandler {
	return &TaskQueueHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *TaskQueueHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}-task-one", h.taskOne)
	group.Add("{{.ModuleName}}-task-two", h.taskTwo)
}

func (h *TaskQueueHandler) taskOne(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryTaskQueue:TaskOne")
	defer trace.Finish()
	ctx = trace.Context()

	retried := candishared.GetValueFromContext(ctx, candishared.ContextKeyTaskQueueRetry).(int)
	fmt.Printf("executing task '{{.ModuleName}}-task-one' has been %d retry\n", retried)
	return &candishared.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error one",
	}
}

func (h *TaskQueueHandler) taskTwo(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryTaskQueue:TaskTwo")
	defer trace.Finish()
	ctx = trace.Context()

	retried := candishared.GetValueFromContext(ctx, candishared.ContextKeyTaskQueueRetry).(int)
	fmt.Printf("executing task '{{.ModuleName}}-task-two' has been %d retry\n", retried)
	return &candishared.ErrorRetrier{
		Delay:   3 * time.Second,
		Message: "Error two",
	}
}
`

	deliveryPostgresListenerTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// PostgresListenerHandler struct
type PostgresListenerHandler struct {
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewPostgresListenerHandler constructor
func NewPostgresListenerHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *PostgresListenerHandler {
	return &PostgresListenerHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *PostgresListenerHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}s", h.handleDataChangeOn{{clean (upper .ModuleName)}}) // listen data change on table "{{.ModuleName}}s"
}

func (h *PostgresListenerHandler) handleDataChangeOn{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryPostgresListener:HandleDataChange{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	fmt.Printf("data change on table {{.ModuleName}}s detected: %s\n", message)
	return nil
}
`

	deliveryRabbitMQTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// RabbitMQHandler struct
type RabbitMQHandler struct {
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewRabbitMQHandler constructor
func NewRabbitMQHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *RabbitMQHandler {
	return &RabbitMQHandler{
		uc:        uc,
		validator: validator,
	}
}

// MountHandlers mount handler group
func (h *RabbitMQHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}", h.handleQueue{{clean (upper .ModuleName)}}) // consume queue "{{.ModuleName}}"
}

func (h *RabbitMQHandler) handleQueue{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryRabbitMQ:HandleQueue{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	fmt.Printf("message consumed by module {{.ModuleName}}. message: %s\n", message)
	return nil
}
`
)
