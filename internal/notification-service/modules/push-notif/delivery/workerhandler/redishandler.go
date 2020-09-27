package workerhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	taskqueueworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/task_queue_worker"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

// RedisHandler struct
type RedisHandler struct {
	uc usecase.PushNotifUsecase
}

// NewRedisHandler constructor
func NewRedisHandler(uc usecase.PushNotifUsecase) *RedisHandler {
	return &RedisHandler{
		uc: uc,
	}
}

// MountHandlers return map topic to handler func
func (h *RedisHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	group.Add("scheduled-push-notif", h.handleScheduledNotif, redisErrorHandler("task-retry-redis-push-notif-error"))
	group.Add("push", h.handlePush)
	group.Add("broadcast-topic", h.publishMessageToTopic)
}

func (h *RedisHandler) handleScheduledNotif(ctx context.Context, message []byte) error {
	var payload domain.PushNotifRequestPayload
	json.Unmarshal(message, &payload)
	err := h.uc.SendNotification(ctx, &payload)
	fmt.Println("mantab")
	return err
}

func (h *RedisHandler) handlePush(ctx context.Context, message []byte) error {
	fmt.Println("check")
	time.Sleep(50 * time.Second) // heavy process
	fmt.Println("check done")
	return nil
}

func (h *RedisHandler) publishMessageToTopic(ctx context.Context, message []byte) error {

	var eventPayload domain.Event
	json.Unmarshal(message, &eventPayload)
	h.uc.PublishMessageToTopic(ctx, &eventPayload)
	return nil
}

func redisErrorHandler(taskName string) types.WorkerErrorHandler {
	return func(ctx context.Context, workerType types.Worker, workerName string, message []byte, err error) {

		logger.LogYellow(string(workerType) + " - " + workerName + " - " + string(message) + " - handling error: " + string(err.Error()))

		// add job in task queue for retry, task must registered in `taskqueuehandler`
		if err := taskqueueworker.AddJob(taskName, 5, message); err != nil {
			logger.LogE(err.Error())
		}
	}
}
