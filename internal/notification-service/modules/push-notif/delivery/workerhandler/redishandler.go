package workerhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

// RedisHandler struct
type RedisHandler struct {
	modName string
	uc      usecase.PushNotifUsecase
}

// NewRedisHandler constructor
func NewRedisHandler(modName types.Module, uc usecase.PushNotifUsecase) *RedisHandler {
	return &RedisHandler{
		modName: string(modName),
		uc:      uc,
	}
}

// MountHandlers return map topic to handler func
func (h *RedisHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		helper.BuildRedisPubSubKeyTopic(h.modName, "scheduled-push-notif"): h.handleScheduledNotif,
		helper.BuildRedisPubSubKeyTopic(h.modName, "push"):                 h.handlePush,
	}
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
