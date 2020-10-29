package types

import "context"

// GraphQLMiddlewareFunc type
type GraphQLMiddlewareFunc func(context.Context) context.Context

// GraphQLMiddlewareGroup type
type GraphQLMiddlewareGroup map[string]GraphQLMiddlewareFunc

// Add register resolver to middleware
func (mw GraphQLMiddlewareGroup) Add(schemaResolverName string, middlewareFunc GraphQLMiddlewareFunc) {
	mw[schemaResolverName] = middlewareFunc
}
