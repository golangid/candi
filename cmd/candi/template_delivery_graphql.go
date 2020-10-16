package main

const (
	deliveryGraphqlRootTemplate = `// {{.Header}}

package graphqlhandler

import (
	"{{.GoModName}}/internal/modules/{{clean .ModuleName}}/usecase"
	
	"{{.PackageName}}/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
	mw interfaces.Middleware
	uc usecase.{{clean (upper .ModuleName)}}Usecase
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware, uc usecase.{{clean (upper .ModuleName)}}Usecase) *GraphQLHandler {

	h := &GraphQLHandler{
		mw: mw, uc: uc,
	}

	return h
}

// Query method
func (h *GraphQLHandler) Query() interface{} {
	return &queryResolver{root: h}
}

// Mutation method
func (h *GraphQLHandler) Mutation() interface{} {
	return &mutationResolver{root: h}
}

// Subscription method
func (h *GraphQLHandler) Subscription() interface{} {
	return &subscriptionResolver{root: h}
}
`

	deliveryGraphqlQueryTemplate = `// {{.Header}}

package graphqlhandler

import (
	"context"

	"{{.PackageName}}/tracer"
)

type queryResolver struct {
	root *GraphQLHandler
}

// Hello resolver
func (q *queryResolver) Hello(ctx context.Context) (string, error) {
	trace := tracer.StartTrace(ctx, "DeliveryGraphQL-Hello")
	defer trace.Finish()
	ctx = trace.Context()

	return q.root.uc.Hello(ctx), nil
}
`
	deliveryGraphqlMutationTemplate = `// {{.Header}}

package graphqlhandler

import "context"

type mutationResolver struct {
	root *GraphQLHandler
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
	root *GraphQLHandler
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
	{{clean $module.ModuleName}}: {{clean (upper $module.ModuleName)}}QueryModule
{{- end }}
}

type Mutation {
{{- range $module := .Modules}}
	{{clean $module.ModuleName}}: {{clean (upper $module.ModuleName)}}MutationModule
{{- end }}
}

type Subscription {
{{- range $module := .Modules}}
	{{clean $module.ModuleName}}: {{clean (upper $module.ModuleName)}}SubscriptionModule
{{- end }}
}
`

	defaultGraphqlSchema = `# {{.Header}}

# {{clean (upper .ModuleName)}}Module Module Area
type {{clean (upper .ModuleName)}}QueryModule {
    hello(): String!
}

type {{clean (upper .ModuleName)}}MutationModule {
    hello(): String!
}

type {{clean (upper .ModuleName)}}SubscriptionModule {
    hello(): String!
}
`
)
