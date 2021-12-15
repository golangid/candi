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
	group.Add("{{.ModuleName}}", h.handle{{upper (camel .ModuleName)}}) // handling topic "{{.ModuleName}}"
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) handle{{upper (camel .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryKafka:Handle{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	key := candishared.GetValueFromContext(ctx, candishared.ContextKeyWorkerKey).([]byte)
	fmt.Printf("message consumed by module {{.ModuleName}}. key: %s, message: %s\n", key, message)
	// exec usecase
	return nil
}
`

	deliveryCronTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	cronworker "github.com/golangid/candi/codebase/app/cron_worker"
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
	group.Add(cronworker.CreateCronJobKey("{{.ModuleName}}-scheduler", "message", "10s"), h.handle{{upper (camel .ModuleName)}})
}

func (h *CronHandler) handle{{upper (camel .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryCron:Handle{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Println("cron: execute in module {{.ModuleName}}, message:", string(message))
	// exec usecase
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
	group.Add("{{.ModuleName}}-sample", h.handle{{upper (camel .ModuleName)}})
}

func (h *RedisHandler) handle{{upper (camel .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryRedis:Handle{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Println("redis subs: execute sample")
	// exec usecase
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
	group.Add("{{.ModuleName}}-task", h.handleTask{{upper (camel .ModuleName)}})
}

func (h *TaskQueueHandler) handleTask{{upper (camel .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryTaskQueue:HandleTask{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	retried := candishared.GetValueFromContext(ctx, candishared.ContextKeyTaskQueueRetry).(int)
	fmt.Printf("executing task '{{.ModuleName}}-task' has been %d retry\n", retried)
	// exec usecase
	return &candishared.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error retry",
	}
}
`

	deliveryPostgresListenerTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"encoding/json"
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/candihelper"
	postgresworker "{{.LibraryName}}/codebase/app/postgres_worker"
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
	group.Add("{{snake .ModuleName}}s", h.handleDataChangeOn{{upper (camel .ModuleName)}}) // listen data change on table "{{.ModuleName}}s"
}

func (h *PostgresListenerHandler) handleDataChangeOn{{upper (camel .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryPostgresListener:HandleDataChange{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	var payload postgresworker.EventPayload
	json.Unmarshal(message, &payload)
	fmt.Printf("data change on table '%s' with action '%s' detected. \nOld values: %s\nNew Values: %s\n",
		payload.Table, payload.Action, candihelper.ToBytes(payload.Data.Old), candihelper.ToBytes(payload.Data.New))
	// exec usecase
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
	group.Add("{{.ModuleName}}", h.handleQueue{{upper (camel .ModuleName)}}) // consume queue "{{.ModuleName}}"
}

func (h *RabbitMQHandler) handleQueue{{upper (camel .ModuleName)}}(ctx context.Context, message []byte) error {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryRabbitMQ:HandleQueue{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("message consumed by module {{.ModuleName}}. message: %s\n", message)
	// exec usecase
	return nil
}
`
)
