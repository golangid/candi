package delivery

import (
	"context"
	"fmt"

	"github.com/agungdwiprasetyo/backend-microservices/pkg/middleware"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/shared"
)

// GraphQLHandler model
type GraphQLHandler struct {
	mw middleware.Middleware
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw middleware.Middleware) *GraphQLHandler {
	return &GraphQLHandler{
		mw: mw,
	}
}

// GetAll handler
func (h *GraphQLHandler) GetAll(ctx context.Context, filter struct{ *shared.Filter }) (string, error) {
	fmt.Printf("%+v\n", filter)
	return "OK", nil
}
