package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

type queryResolver struct {
	uc usecase.PushNotifUsecase
	mw interfaces.Middleware
}

// Hello resolver
func (q *queryResolver) Hello(ctx context.Context) (string, error) {
	trace := tracer.StartTrace(ctx, "Delivery-Hello")
	defer trace.Finish()

	q.mw.GraphQLBasicAuth(ctx)

	return "Hello, from service: notification-service, module: push-notif", nil
}
