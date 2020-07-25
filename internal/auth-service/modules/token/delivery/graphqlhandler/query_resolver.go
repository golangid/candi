package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

type queryResolver struct {
}

// Hello resolver
func (q *queryResolver) Hello(ctx context.Context) (string, error) {
	trace := tracer.StartTrace(ctx, "Delivery-Hello")
	defer trace.Finish()

	return "Hello, from service: auth-service, module: token", nil
}
