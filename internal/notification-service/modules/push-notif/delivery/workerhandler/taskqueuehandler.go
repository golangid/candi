package workerhandler

import (
	"context"
	"encoding/json"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	taskqueueworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/task_queue_worker"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

// TaskQueueHandler struct
type TaskQueueHandler struct {
	uc usecase.PushNotifUsecase
}

// NewTaskQueueHandler constructor
func NewTaskQueueHandler(uc usecase.PushNotifUsecase) *TaskQueueHandler {
	return &TaskQueueHandler{
		uc: uc,
	}
}

// MountHandlers return map topic to handler func
func (h *TaskQueueHandler) MountHandlers(group *types.WorkerHandlerGroup) {

	group.Add("task-one", h.taskOne)
	group.Add("task-two", h.taskTwo)
}

func (h *TaskQueueHandler) taskOne(ctx context.Context, message []byte) error {
	logger.LogRed(time.Now().String() + " task-one: " + string(message))
	return &taskqueueworker.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error gan",
	}
}

func (h *TaskQueueHandler) taskTwo(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "DeliveryTaskQueue:TaskTwo")
	defer trace.Finish()

	var mm map[string]string
	json.Unmarshal(message, &mm)

	logger.LogYellow("task-two: " + string(message))
	return &taskqueueworker.ErrorRetrier{
		Delay:   3 * time.Second,
		Message: "Error gan",
	}
}
