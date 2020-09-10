package workerhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
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
	group.Add("scheduled-push-notif", h.handleScheduledNotif)
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
