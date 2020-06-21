package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware) *GraphQLHandler {
	return &GraphQLHandler{}
}

// Hello resolver
func (h *GraphQLHandler) Hello(ctx context.Context) (string, error) {
	return "Hello, from service: auth-service, module: token", nil
}

