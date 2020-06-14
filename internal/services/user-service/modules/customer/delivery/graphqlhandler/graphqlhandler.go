package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
)

// GraphQLHandler model
type GraphQLHandler struct {
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw middleware.Middleware) *GraphQLHandler {
	return &GraphQLHandler{}
}

// Hello resolver
func (h *GraphQLHandler) Hello(ctx context.Context) (string, error) {
	return "Hello, from service: user-service, module: customer", nil
}

