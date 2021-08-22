package main

const (
	deliveryKafkaTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// KafkaHandler struct
type KafkaHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewKafkaHandler constructor
func NewKafkaHandler(uc usecase.Usecase, deps dependency.Dependency) *KafkaHandler {
	return &KafkaHandler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *KafkaHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}", h.handle{{clean (upper .ModuleName)}}) // handling topic "{{.ModuleName}}"
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) handle{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryKafka:Hello")
	defer trace.Finish()

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

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// CronHandler struct
type CronHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewCronHandler constructor
func NewCronHandler(uc usecase.Usecase, deps dependency.Dependency) *CronHandler {
	return &CronHandler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *CronHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add(candihelper.CronJobKeyToString("{{.ModuleName}}-scheduler", "message", "10s"), h.handle{{clean (upper .ModuleName)}})
}

func (h *CronHandler) handle{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryCron:Hello")
	defer trace.Finish()

	fmt.Println("cron: execute in module {{.ModuleName}}, message:", string(message))
	return nil
}
`

	deliveryRedisTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// RedisHandler struct
type RedisHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewRedisHandler constructor
func NewRedisHandler(uc usecase.Usecase, deps dependency.Dependency) *RedisHandler {
	return &RedisHandler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *RedisHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}-sample", h.handle{{clean (upper .ModuleName)}})
}

func (h *RedisHandler) handle{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryRedis:Hello")
	defer trace.Finish()

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

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// TaskQueueHandler struct
type TaskQueueHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewTaskQueueHandler constructor
func NewTaskQueueHandler(uc usecase.Usecase, deps dependency.Dependency) *TaskQueueHandler {
	return &TaskQueueHandler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *TaskQueueHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}-task-one", h.taskOne)
	group.Add("{{.ModuleName}}-task-two", h.taskTwo)
}

func (h *TaskQueueHandler) taskOne(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryTaskQueue:TaskOne")
	defer trace.Finish()

	retried := candishared.GetValueFromContext(ctx, candishared.ContextKeyTaskQueueRetry).(int)
	fmt.Printf("executing task '{{.ModuleName}}-task-one' has been %d retry\n", retried)
	return &candishared.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error one",
	}
}

func (h *TaskQueueHandler) taskTwo(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryTaskQueue:TaskTwo")
	defer trace.Finish()

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

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// PostgresListenerHandler struct
type PostgresListenerHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewPostgresListenerHandler constructor
func NewPostgresListenerHandler(uc usecase.Usecase, deps dependency.Dependency) *PostgresListenerHandler {
	return &PostgresListenerHandler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *PostgresListenerHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}s", h.handleDataChangeOn{{clean (upper .ModuleName)}}) // listen data change on table "{{.ModuleName}}s"
}

func (h *PostgresListenerHandler) handleDataChangeOn{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryPostgresListener:HandleDataChange{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("data change on table {{.ModuleName}}s detected: %s\n", message)
	return nil
}
`

	deliveryRabbitMQTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// RabbitMQHandler struct
type RabbitMQHandler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewRabbitMQHandler constructor
func NewRabbitMQHandler(uc usecase.Usecase, deps dependency.Dependency) *RabbitMQHandler {
	return &RabbitMQHandler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *RabbitMQHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}", h.handleQueue{{clean (upper .ModuleName)}}) // consume queue "{{.ModuleName}}"
}

func (h *RabbitMQHandler) handleQueue{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryRabbitMQ:HandleQueue{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("message consumed by module {{.ModuleName}}. message: %s\n", message)
	return nil
}
`
)
