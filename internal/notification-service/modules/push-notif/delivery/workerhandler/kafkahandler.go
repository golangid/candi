package workerhandler

import (
	"context"
	"encoding/json"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
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
func (h *KafkaHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		"push-notif":      h.handlePushNotif,
		"notif-broadcast": h.handleNotifBroadcast,
	}
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
