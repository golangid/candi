package graphqlhandler

import (
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
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
func NewGraphQLHandler(rootName string, mw interfaces.Middleware, uc usecase.PushNotifUsecase) *GraphQLHandler {

	h := &GraphQLHandler{
		rootName:     rootName,
		query:        &queryResolver{uc, mw},
		mutation:     &mutationResolver{uc, mw},
		subscription: &subscriptionResolver{uc, mw},
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
	return h.mutation
}

// Subscription method
func (h *GraphQLHandler) Subscription() interface{} {
	return h.subscription
}
