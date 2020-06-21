package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// GraphQLHandler model
type GraphQLHandler struct {
	uc usecase.EventUsecase
	mw interfaces.Middleware
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware, uc usecase.EventUsecase) *GraphQLHandler {
	return &GraphQLHandler{
		uc: uc,
		mw: mw,
	}
}

// GetAll handler
func (h *GraphQLHandler) GetAll(ctx context.Context, filter struct{ *shared.Filter }) (*EventListResolver, error) {
	h.mw.GraphQLBasicAuth(ctx)

	events, meta, err := h.uc.FindAll(ctx, filter.Filter)
	if err != nil {
		return nil, err
	}

	var eventResolvers []*EventResolver
	for _, event := range events {
		eventResolvers = append(eventResolvers, &EventResolver{
			e: event,
			message: EventMessage{
				e: event,
			},
		})
	}

	resolvers := EventListResolver{
		m:      meta,
		events: eventResolvers,
	}
	return &resolvers, nil
}
