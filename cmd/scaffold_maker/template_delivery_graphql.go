package main

const (
	deliveryGraphqlTemplate = `package graphqlhandler

import (
	"context"

	"{{.PackageName}}/pkg/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware) *GraphQLHandler {
	return &GraphQLHandler{}
}

// Hello resolver
func (h *GraphQLHandler) Hello(ctx context.Context) (string, error) {
	return "Hello, from service: {{$.ServiceName}}, module: {{$.module}}", nil
}

`

	defaultGraphqlRootSchema = `schema {
	query: Query
	mutation: Mutation
}

type Query {
{{- range $module := .Modules}}
	{{clean $module}}: {{clean (upper $module)}}Module
{{- end }}
}

type Mutation {
{{- range $module := .Modules}}
	{{clean $module}}: {{clean (upper $module)}}Module
{{- end }}
}
`

	defaultGraphqlSchema = `################### {{clean (upper $.module)}}Module Module Area
type {{clean (upper $.module)}}Module {
    hello(): String!
}
`
)
