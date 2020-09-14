package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	taskqueueworker "agungdwiprasetyo.com/backend-microservices/pkg/codebase/app/task_queue_worker"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

type queryResolver struct {
	uc usecase.PushNotifUsecase
	mw interfaces.Middleware
}

// Hello resolver
func (q *queryResolver) Hello(ctx context.Context, input struct {
	JobName  string
	Args     string
	MaxRetry int32
}) (string, error) {
	trace := tracer.StartTrace(ctx, "Delivery-Hello")
	defer trace.Finish()

	// q.mw.GraphQLBasicAuth(ctx)
	if err := taskqueueworker.AddJob(input.JobName, int(input.MaxRetry), helper.ToBytes(input.Args)); err != nil {
		return "", shared.NewGraphQLErrorResolver(err.Error(), map[string]interface{}{
			"code": "BadRequest",
		})
	}

	return "Hello, from service: notification-service, module: push-notif", nil
}
