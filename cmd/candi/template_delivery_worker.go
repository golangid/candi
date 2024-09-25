package main

import (
	"strings"

	"github.com/golangid/candi/candihelper"
)

const (
	deliveryKafkaTemplate = `// {{.Header}}

package workerhandler

import (
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/tracer"
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
func (h *KafkaHandler) handle{{upper (camel .ModuleName)}}(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "{{upper (camel .ModuleName)}}DeliveryKafka:Handle{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("message consumed in handler %s. key: %s, message: %s\n", eventContext.HandlerRoute(), eventContext.Key(), eventContext.Message())

	// exec usecase
	// h.uc.SomethingUsecase()

	return ctx.Err()
}
`

	deliveryCronTemplate = `// {{.Header}}

package workerhandler

import (
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/candishared"
	cronworker "github.com/golangid/candi/codebase/app/cron_worker"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/tracer"
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
	group.Add(cronworker.CreateCronJobKey("{{.ModuleName}}-scheduler", "message", "* * * * *"), h.handle{{upper (camel .ModuleName)}})
}

func (h *CronHandler) handle{{upper (camel .ModuleName)}}(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "{{upper (camel .ModuleName)}}DeliveryCron:Handle{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("cron: execute in handler %s, message: %s\n", eventContext.HandlerRoute(), eventContext.Message())

	// exec usecase
	// h.uc.SomethingUsecase()

	return ctx.Err()
}
`

	deliveryRedisTemplate = `// {{.Header}}

package workerhandler

import (
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/tracer"
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

func (h *RedisHandler) handle{{upper (camel .ModuleName)}}(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "{{upper (camel .ModuleName)}}DeliveryRedis:Handle{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("redis subs: execute handler %s with message %s\n", eventContext.HandlerRoute(), eventContext.Message())

	// exec usecase
	// h.uc.SomethingUsecase()

	return ctx.Err()
}
`

	deliveryTaskQueueTemplate = `// {{.Header}}

package workerhandler

import (
	"fmt"
	"time"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/candishared"
	taskqueueworker "github.com/golangid/candi/codebase/app/task_queue_worker"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/tracer"
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
	group.Add("{{.ModuleName}}-task", h.handleTask{{upper (camel .ModuleName)}},
		types.WorkerHandlerOptionAddConfig(taskqueueworker.TaskOptionDeleteJobAfterSuccess, false),
	)
}

func (h *TaskQueueHandler) handleTask{{upper (camel .ModuleName)}}(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "{{upper (camel .ModuleName)}}DeliveryTaskQueue:HandleTask{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("executing task '%s' has been %s retry, with message: %s\n",
		eventContext.HandlerRoute(),
		eventContext.Header()["retries"],
		eventContext.Message(),
	)

	// exec usecase
	// h.uc.SomethingUsecase()
	time.Sleep(1 * time.Second) // just for example process
	_ = ctx.Err()

	return &candishared.ErrorRetrier{
		Delay:   1 * time.Second,
		Message: "Error retry",
	}
}
`

	deliveryPostgresListenerTemplate = `// {{.Header}}

package workerhandler

import (
	"encoding/json"
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	postgresworker "github.com/golangid/candi/codebase/app/postgres_worker"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/tracer"
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
	group.Add("{{plural .ModuleName}}", h.handleDataChangeOn{{upper (camel .ModuleName)}}) // listen data change on table "{{plural .ModuleName}}"
}

func (h *PostgresListenerHandler) handleDataChangeOn{{upper (camel .ModuleName)}}(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "{{upper (camel .ModuleName)}}DeliveryPostgresListener:HandleDataChange{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	var payload postgresworker.EventPayload
	json.Unmarshal(eventContext.Message(), &payload)
	fmt.Printf("data change on table '%s' with action '%s' detected. \nOld values: %s\nNew Values: %s\n",
		payload.Table, payload.Action, candihelper.ToBytes(payload.Data.Old), candihelper.ToBytes(payload.Data.New))
	if payload.Data.IsTooLongPayload {
		fmt.Printf("Data too long! \nold id: %s\nnew id: %s\n", payload.Data.OldID, payload.Data.NewID)
	}

	// exec usecase
	// h.uc.SomethingUsecase()

	return ctx.Err()
}
`

	deliveryRabbitMQTemplate = `// {{.Header}}

package workerhandler

import (
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/tracer"
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

func (h *RabbitMQHandler) handleQueue{{upper (camel .ModuleName)}}(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "{{upper (camel .ModuleName)}}DeliveryRabbitMQ:HandleQueue{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("message consumed by module {{.ModuleName}}. message: %s\n", eventContext.Message())

	// exec usecase
	// h.uc.SomethingUsecase()

	return ctx.Err()
}
`

	deliveryWorkerPluginTemplate = `// {{.Header}}

package workerhandler

import (
	"fmt"

	"{{.PackagePrefix}}/pkg/shared/usecase"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/tracer"
)

// {{.WorkerPluginName}}Handler struct
type {{.WorkerPluginName}}Handler struct {
	uc        usecase.Usecase
	validator interfaces.Validator
}

// New{{.WorkerPluginName}}Handler constructor
func New{{.WorkerPluginName}}Handler(uc usecase.Usecase, deps dependency.Dependency) *{{.WorkerPluginName}}Handler {
	return &{{.WorkerPluginName}}Handler{
		uc:        uc,
		validator: deps.GetValidator(),
	}
}

// MountHandlers mount handler group
func (h *{{.WorkerPluginName}}Handler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}", h.handleTopic{{upper (camel .ModuleName)}}) // consume topic "{{.ModuleName}}"
}

func (h *{{.WorkerPluginName}}Handler) handleTopic{{upper (camel .ModuleName)}}(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "{{upper (camel .ModuleName)}}Delivery{{.WorkerPluginName}}:HandleTopic{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	fmt.Printf("message consumed by module {{.ModuleName}}. message: %s\n", eventContext.Context())

	// exec usecase
	// h.uc.SomethingUsecase()

	return ctx.Err()
}
`
)

func getWorkerFuncTemplate(workerName, moduleName, usecaseName string) string {
	moduleName, usecaseName = strings.Title(moduleName), strings.Title(usecaseName)
	return `func (h *` + workerName + `Handler) ` + candihelper.ToCamelCase(usecaseName) + `(eventContext *candishared.EventContext) error {
	trace, ctx := tracer.StartTraceWithContext(eventContext.Context(), "` + moduleName + `Delivery` + workerName + `:` + usecaseName + `")
	defer trace.Finish()

	var payload domain.Request` + usecaseName + `
	if err := json.Unmarshal(eventContext.Message(), &payload); err != nil {
		return err
	}
	_, err := h.uc.` + moduleName + `().` + usecaseName + `(ctx, &payload)
	return err
}
`
}
