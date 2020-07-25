package workerhandler

import (
	"context"
	"fmt"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

// CronHandler struct
type CronHandler struct {
	uc usecase.PushNotifUsecase
}

// NewCronHandler constructor
func NewCronHandler(uc usecase.PushNotifUsecase) *CronHandler {
	return &CronHandler{
		uc: uc,
	}
}

// MountHandlers return map topic to handler func
func (h *CronHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		helper.CronJobKeyToString("push-notif", "00:00:00"): h.handleScheduledMidnight,
		helper.CronJobKeyToString("push", "10s"):            h.handleCheck,
	}
}

func (h *CronHandler) handleScheduledMidnight(ctx context.Context, message []byte) error {
	fmt.Println("execute scheduled midnight")
	return nil
}

func (h *CronHandler) handleCheck(ctx context.Context, message []byte) error {
	fmt.Println("check")
	time.Sleep(50 * time.Second) // heavy process
	fmt.Println("check done")
	return nil
}
