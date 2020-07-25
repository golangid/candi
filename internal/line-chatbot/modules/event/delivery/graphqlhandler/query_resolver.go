package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

type queryResolver struct {
	uc usecase.EventUsecase
	mw interfaces.Middleware
}

// Hello resolver
func (q *queryResolver) GetAll(ctx context.Context, filter struct{ *shared.Filter }) (*EventListResolver, error) {
	q.mw.GraphQLBasicAuth(ctx)

	events, meta, err := q.uc.FindAll(ctx, filter.Filter)
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
