package workerhandler

import (
	"context"
	"encoding/json"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	taskqueueworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/task_queue_worker"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

// KafkaHandler struct
type KafkaHandler struct {
	uc usecase.PushNotifUsecase
}

// NewKafkaHandler constructor
func NewKafkaHandler(uc usecase.PushNotifUsecase) *KafkaHandler {
	return &KafkaHandler{
		uc: uc,
	}
}

// MountHandlers return map topic to handler func
func (h *KafkaHandler) MountHandlers(group *types.WorkerHandlerGroup) {

	group.Add("push-notif", h.handlePushNotif, kafkaTopicErrorHandler("task-retry-kafka-push-notif-error"))
	group.Add("notif-broadcast", h.handleNotifBroadcast, kafkaTopicErrorHandler("task-two"))
}

func (h *KafkaHandler) handlePushNotif(ctx context.Context, message []byte) (err error) {

	var payload domain.PushNotifRequestPayload
	json.Unmarshal(message, &payload)
	err = h.uc.SendNotification(ctx, &payload)

	if err != nil {
		logger.LogE(err.Error())
	}

	return err
}

func (h *KafkaHandler) handleNotifBroadcast(ctx context.Context, message []byte) (err error) {
	var payload domain.Event
	json.Unmarshal(message, &payload)
	h.uc.PublishMessageToTopic(ctx, &payload)

	return
}

func kafkaTopicErrorHandler(taskName string) types.WorkerErrorHandler {
	return func(ctx context.Context, workerType types.Worker, workerName string, message []byte, err error) {

		logger.LogYellow(string(workerType) + " - " + workerName + " - " + string(message) + " - handling error: " + string(err.Error()))

		// add job in task queue for retry, task must registered in `taskqueuehandler`
		if err := taskqueueworker.AddJob(taskName, 5, message); err != nil {
			logger.LogE(err.Error())
		}
	}
}
