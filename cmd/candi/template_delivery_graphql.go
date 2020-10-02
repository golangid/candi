package main

const (
	deliveryGraphqlRootTemplate = `// {{.Header}}

package graphqlhandler

import (
	"{{.PackageName}}/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
	query        *queryResolver
	mutation     *mutationResolver
	subscription *subscriptionResolver
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware) *GraphQLHandler {

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
`

	deliveryGraphqlQueryTemplate = `// {{.Header}}

package graphqlhandler

import (
	"context"

	"{{.PackageName}}/tracer"
)

type queryResolver struct {
}

// Hello resolver
func (q *queryResolver) Hello(ctx context.Context) (string, error) {
	trace := tracer.StartTrace(ctx, "Delivery-Hello")
	defer trace.Finish()

	return "Hello, from service: {{$.ServiceName}}, module: {{$.module}}", nil
}
`
	deliveryGraphqlMutationTemplate = `// {{.Header}}

package graphqlhandler

import "context"

type mutationResolver struct {
}

// Hello resolver
func (m *mutationResolver) Hello(ctx context.Context) (string, error) {
	return "Hello", nil
}	
`
	deliveryGraphqlSubscriptionTemplate = `// {{.Header}}

package graphqlhandler

import "context"

type subscriptionResolver struct {
}

// Hello resolver
func (s *subscriptionResolver) Hello(ctx context.Context) <-chan string {
	output := make(chan string)

	go func() {
		output <- "Hello"
	}()

	return output
}
`

	defaultGraphqlRootSchema = `# {{.Header}}

schema {
	query: Query
	mutation: Mutation
	subscription: Subscription
}

type Query {
{{- range $module := .Modules}}
	{{clean $module.Name}}: {{clean (upper $module.Name)}}QueryModule
{{- end }}
}

type Mutation {
{{- range $module := .Modules}}
	{{clean $module.Name}}: {{clean (upper $module.Name)}}MutationModule
{{- end }}
}

type Subscription {
{{- range $module := .Modules}}
	{{clean $module.Name}}: {{clean (upper $module.Name)}}SubscriptionModule
{{- end }}
}
`

	defaultGraphqlSchema = `# {{.Header}}

# {{clean (upper $.module)}}Module Module Area
type {{clean (upper $.module)}}QueryModule {
    hello(): String!
}

type {{clean (upper $.module)}}MutationModule {
    hello(): String!
}

type {{clean (upper $.module)}}SubscriptionModule {
    hello(): String!
}
`
)
