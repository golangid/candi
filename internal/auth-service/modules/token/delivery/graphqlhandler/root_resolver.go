package graphqlhandler

import (
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
	query        *queryResolver
	mutation     *mutationResolver
	subscription *subscriptionResolver
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware, uc usecase.TokenUsecase) *GraphQLHandler {

	h := &GraphQLHandler{
		query:        &queryResolver{},
		mutation:     &mutationResolver{},
		subscription: &subscriptionResolver{},
	}

	return h
}

// Query method
func (h *GraphQLHandler) Query() interface{} {
	return h.query
}

// Mutation method
func (h *GraphQLHandler) Mutation() interface{} {
	return h.mutation
}

// Subscription method
func (h *GraphQLHandler) Subscription() interface{} {
	return h.subscription
}
