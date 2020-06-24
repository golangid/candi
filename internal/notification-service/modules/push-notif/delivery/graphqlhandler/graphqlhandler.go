package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/utils"
)

// GraphQLHandler model
type GraphQLHandler struct {
	mw interfaces.Middleware
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware) *GraphQLHandler {
	return &GraphQLHandler{
		mw: mw,
	}
}

// Hello resolver
func (h *GraphQLHandler) Hello(ctx context.Context) (string, error) {
	trace := utils.StartTrace(ctx, "Delivery-PushNotif")
	defer trace.Finish()

	h.mw.GraphQLBasicAuth(ctx)

	return "Hello, from service: notification-service, module: push-notif", nil
}
