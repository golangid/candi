package graphqlhandler

import (
	"context"
	"errors"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/utils"
)

// GraphQLHandler model
type GraphQLHandler struct {
	mw interfaces.Middleware
	uc usecase.PushNotifUsecase
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware, uc usecase.PushNotifUsecase) *GraphQLHandler {
	return &GraphQLHandler{
		mw: mw,
		uc: uc,
	}
}

// Hello resolver
func (h *GraphQLHandler) Hello(ctx context.Context) (string, error) {
	trace := utils.StartTrace(ctx, "Delivery-Hello")
	defer trace.Finish()

	h.mw.GraphQLBasicAuth(ctx)

	return "Hello, from service: notification-service, module: push-notif", nil
}

// Push resolver
func (h *GraphQLHandler) Push(ctx context.Context, input pushInputResolver) (string, error) {
	trace := utils.StartTrace(ctx, "Delivery-Push")
	defer trace.Finish()
	h.mw.GraphQLBasicAuth(ctx)

	ctx, tags := trace.Context(), trace.Tags()
	tags["input"] = input

	if err := h.uc.SendNotification(ctx, input.Payload); err != nil {
		return "", err
	}

	return "Ok", nil
}

// ScheduledNotification resolver
func (h *GraphQLHandler) ScheduledNotification(ctx context.Context, input scheduleNotifInputResolver) (string, error) {
	h.mw.GraphQLBasicAuth(ctx)

	scheduledAt, err := time.Parse(time.RFC3339, input.Payload.ScheduledAt)
	if err != nil {
		return "Failed parse scheduled time format", err
	}

	if scheduledAt.Before(time.Now()) {
		return "", errors.New("Scheduled time must in future")
	}

	if err := h.uc.SendScheduledNotification(ctx, scheduledAt, input.Payload.Data); err != nil {
		return "Failed set scheduled push notification", err
	}

	return "Success set scheduled push notification, scheduled at: " + input.Payload.ScheduledAt, nil
}
