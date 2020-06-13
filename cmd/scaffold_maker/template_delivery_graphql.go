package main

const (
	deliveryGraphqlTemplate = `package graphqlhandler

import (
	"context"

	"{{.PackageName}}/pkg/middleware"
)

// GraphQLHandler model
type GraphQLHandler struct {
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw middleware.Middleware) *GraphQLHandler {
	return &GraphQLHandler{}
}

// Hello resolver
func (h *GraphQLHandler) Hello(ctx context.Context) (string, error) {
	return "Hello, from service: {{$.ServiceName}}, module: {{$.module}}", nil
}

`

	defaultGraphqlSchema = `
schema {
	query: Query
	mutation: Mutation
}

type Query {
{{- range $module := .Modules}}
	{{$module}}: {{upper $module}}Module
{{- end }}
}

type Mutation {
{{- range $module := .Modules}}
	{{$module}}: {{upper $module}}Module
{{- end }}
}

{{- range $module := .Modules}}

################### {{upper $module}}Module Module Area
type {{upper $module}}Module {
    hello(): String!
}

{{- end }}
`
)
