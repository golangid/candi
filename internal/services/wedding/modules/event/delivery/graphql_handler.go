package delivery

import (
	"context"
	"fmt"

	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// GraphQLHandler model
type GraphQLHandler struct {
	mw middleware.Middleware
	uc usecase.EventUsecase
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw middleware.Middleware, uc usecase.EventUsecase) *GraphQLHandler {
	return &GraphQLHandler{
		mw: mw,
		uc: uc,
	}
}

// GetAll handler
func (h *GraphQLHandler) GetAll(ctx context.Context, filter struct{ *shared.Filter }) (string, error) {
	fmt.Printf("%+v\n", filter)
	return "OK", nil
}

// GetByCode handler
func (h *GraphQLHandler) GetByCode(ctx context.Context, args struct{ Code string }) (*EventResolver, error) {
	event, err := h.uc.FindByCode(ctx, args.Code)
	if err != nil {
		return nil, err
	}
	return &EventResolver{e: event}, nil
}
