package graphqlhandler

import (
	"context"
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// GraphQLHandler model
type GraphQLHandler struct {
	uc        usecase.EventUsecase
	basicAuth func(ctx context.Context)
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw middleware.Middleware, uc usecase.EventUsecase) *GraphQLHandler {
	return &GraphQLHandler{
		uc: uc,
		basicAuth: func(ctx context.Context) {
			headers := ctx.Value(shared.ContextKey("headers")).(http.Header)
			if err := mw.BasicAuth(headers.Get("Authorization")); err != nil {
				panic(err)
			}
		},
	}
}

// GetAll handler
func (h *GraphQLHandler) GetAll(ctx context.Context, filter struct{ *shared.Filter }) (*EventListResolver, error) {
	h.basicAuth(ctx)
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
