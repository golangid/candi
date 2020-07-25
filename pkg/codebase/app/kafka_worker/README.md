# Example

```go
package workerhandler

import (
	"context"
	"encoding/json"

	"example.service/internal/modules/push-notif/domain"
	"example.service/internal/modules/push-notif/usecase"
	
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
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

// MountHandlers return group map topic to handler func
func (h *KafkaHandler) MountHandlers() map[string]types.WorkerHandlerFunc {
	return map[string]types.WorkerHandlerFunc{
		"push-notif": h.handlePushNotif,
	}
}

func (h *KafkaHandler) handlePushNotif(ctx context.Context, message []byte) error {
	trace := tracer.StartTrace(ctx, "KafkaDelivery-HandlePushNotif")
	defer trace.Finish()

	var payload domain.PushNotifRequestPayload
	json.Unmarshal(message, &payload)
	return h.uc.SendNotification(ctx, &payload)
}
```