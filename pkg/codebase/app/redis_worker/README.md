# Example

```go
package workerhandler

import (
	"context"
	"encoding/json"
	"time"

	"example.service/internal/modules/push-notif/domain"
	"example.service/internal/modules/push-notif/usecase"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

// RedisHandler struct
type RedisHandler struct {
	uc usecase.PushNotifUsecase
}

// NewRedisHandler constructor
func NewRedisHandler(modName types.Module, uc usecase.PushNotifUsecase) *RedisHandler {
	return &RedisHandler{
		uc: uc,
	}
}

// MountHandlers return group map topic key to handler func
func (h *RedisHandler) MountHandlers() map[string]types.WorkerHandlerFunc {
	return map[string]types.WorkerHandlerFunc{
		"scheduled-push-notif": h.handleScheduledPushNotif,
		"heavy-push":           h.handleHeavyPush,
	}
}

func (h *RedisHandler) handleScheduledPushNotif(ctx context.Context, message []byte) error {
	var payload domain.PushNotifRequestPayload
	json.Unmarshal(message, &payload)
	err := h.uc.SendNotification(ctx, &payload)
	logger.LogIf("success handling message: %s", string(message))
	return err
}

func (h *RedisHandler) handleHeavyPush(ctx context.Context, message []byte) error {
	logger.LogI("start heavy push")
	time.Sleep(30 * time.Second)
	logger.LogI("heavy push done")
	return nil
}
```