package graphqlhandler

import (
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
	rootName     string
	query        *queryResolver
	mutation     *mutationResolver
	subscription *subscriptionResolver
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(rootName string, mw interfaces.Middleware, uc usecase.EventUsecase) *GraphQLHandler {

	h := &GraphQLHandler{
		rootName:     rootName,
		query:        &queryResolver{uc, mw},
		mutation:     &mutationResolver{},
		subscription: &subscriptionResolver{},
	}

	return h
}

// RootName resolver field
func (h *GraphQLHandler) RootName() string {
	return h.rootName
}

// Query method
func (h *GraphQLHandler) Query() interface{} {
	return h.query
}

// Mutation method
func (h *GraphQLHandler) Mutation() interface{} {
	return nil
}

// Subscription method
func (h *GraphQLHandler) Subscription() interface{} {
	return nil
}
