package main

const (
	deliveryKafkaTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.GoModName}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.PackageName}}/codebase/factory/types"
	"{{.PackageName}}/tracer"
)

// KafkaHandler struct
type KafkaHandler struct {
	uc usecase.{{clean (upper .ModuleName)}}Usecase
}

// NewKafkaHandler constructor
func NewKafkaHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase) *KafkaHandler {
	return &KafkaHandler{
		uc: uc,
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

	fmt.Printf("message consumed by module {{.ModuleName}}. message: %s\n", string(message))
	h.uc.Hello(ctx)
	return nil
}
`

	deliveryCronTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.GoModName}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.PackageName}}/candihelper"
	"{{.PackageName}}/codebase/factory/types"
	"{{.PackageName}}/tracer"
)

// CronHandler struct
type CronHandler struct {
	uc usecase.{{clean (upper .ModuleName)}}Usecase
}

// NewCronHandler constructor
func NewCronHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase) *CronHandler {
	return &CronHandler{
		uc: uc,
	}
}

// MountHandlers mount handler group
func (h *CronHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add(candihelper.CronJobKeyToString("{{.ModuleName}}-scheduler", "10s"), h.handle{{clean (upper .ModuleName)}})
}

func (h *CronHandler) handle{{clean (upper .ModuleName)}}(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryCron:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	fmt.Println("cron: execute in module {{.ModuleName}}")
	h.uc.Hello(ctx)
	return nil
}
`

	deliveryRedisTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"fmt"

	"{{.GoModName}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"{{.PackageName}}/codebase/factory/types"
	"{{.PackageName}}/tracer"
)

// RedisHandler struct
type RedisHandler struct {
	uc usecase.{{clean (upper .ModuleName)}}Usecase
}

// NewRedisHandler constructor
func NewRedisHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase) *RedisHandler {
	return &RedisHandler{
		uc: uc,
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
	h.uc.Hello(ctx)
	return nil
}
`

	deliveryTaskQueueTemplate = `// {{.Header}}

package workerhandler

import (
	"context"
	"time"

	"{{.GoModName}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	taskqueueworker "{{.PackageName}}/codebase/app/task_queue_worker"
	"{{.PackageName}}/codebase/factory/types"
	"{{.PackageName}}/tracer"
)

// TaskQueueHandler struct
type TaskQueueHandler struct {
	uc usecase.{{clean (upper .ModuleName)}}Usecase
}

// NewTaskQueueHandler constructor
func NewTaskQueueHandler(uc usecase.{{clean (upper .ModuleName)}}Usecase) *TaskQueueHandler {
	return &TaskQueueHandler{
		uc: uc,
	}
}

// MountHandlers mount handler group
func (h *TaskQueueHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("{{.ModuleName}}-task-one", h.taskOne)
	group.Add("{{.ModuleName}}-task-two", h.taskTwo)
}

func (h *TaskQueueHandler) taskOne(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryTaskQueue:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	h.uc.Hello(ctx)

	return &taskqueueworker.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error one",
	}
}

func (h *TaskQueueHandler) taskTwo(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryTaskQueue:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	h.uc.Hello(ctx)

	return &taskqueueworker.ErrorRetrier{
		Delay:   3 * time.Second,
		Message: "Error two",
	}
}
`
)
