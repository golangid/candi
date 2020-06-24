package graphqlhandler

import (
	"context"

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
