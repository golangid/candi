package workerhandler

import (
	"context"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	taskqueueworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/task_queue_worker"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
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
func (h *TaskQueueHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		"task-one": h.taskOne,
		"task-two": h.taskTwo,
	}
}

func (h *TaskQueueHandler) taskOne(ctx context.Context, message []byte) error {
	logger.LogRed(time.Now().String() + " task-one: " + string(message))
	return &taskqueueworker.ErrorRetrier{
		Delay:   10 * time.Second,
		Message: "Error gan",
	}
}

func (h *TaskQueueHandler) taskTwo(ctx context.Context, message []byte) error {
	logger.LogYellow("task-two: " + string(message))
	return &taskqueueworker.ErrorRetrier{
		Delay:   3 * time.Second,
		Message: "Error gan",
	}
}
