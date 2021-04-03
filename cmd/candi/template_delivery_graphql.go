package main

const (
	deliveryGraphqlRootTemplate = `// {{.Header}}

package graphqlhandler

import (
	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"
	
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
	mw        interfaces.Middleware
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw interfaces.Middleware, uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *GraphQLHandler {
	return &GraphQLHandler{
		mw: mw, uc: uc, validator: validator,
	}
}

// RegisterMiddleware register resolver based on schema in "api/graphql/*" path
func (h *GraphQLHandler) RegisterMiddleware(mwGroup *types.MiddlewareGroup) {
	mwGroup.Add("{{clean (upper .ModuleName)}}QueryResolver.get_all_{{clean .ModuleName}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{clean (upper .ModuleName)}}QueryResolver.get_detail_{{clean .ModuleName}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{clean (upper .ModuleName)}}MutationResolver.save_{{clean .ModuleName}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("{{clean .ModuleName}}.save_{{clean .ModuleName}}"))
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
	
	shareddomain "{{.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/tracer"
)

type queryResolver struct {
	root *GraphQLHandler
}

// GetAll{{clean (upper .ModuleName)}} resolver
func (q *queryResolver) GetAll{{clean (upper .ModuleName)}}(ctx context.Context, input struct{ Filter *CommonFilter }) (results {{clean (upper .ModuleName)}}ListResolver, err error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if input.Filter == nil {
		input.Filter = new(CommonFilter)
	}
	filter := input.Filter.toSharedFilter()
	data, meta, err := q.root.uc.GetAll{{clean (upper .ModuleName)}}(ctx, filter)
	if err != nil {
		return results, err
	}

	return {{clean (upper .ModuleName)}}ListResolver{
		Meta: meta, Data: data,
	}, nil
}

// GetDetail{{clean (upper .ModuleName)}} resolver
func (q *queryResolver) GetDetail{{clean (upper .ModuleName)}}(ctx context.Context, input struct{ ID string }) (data shareddomain.{{clean (upper .ModuleName)}}, err error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:GetDetail{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	return q.root.uc.GetDetail{{clean (upper .ModuleName)}}(ctx, input.ID)
}

`
	deliveryGraphqlMutationTemplate = `// {{.Header}}

package graphqlhandler

import (
	"context"
	
	shareddomain "{{.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/tracer"
)

type mutationResolver struct {
	root *GraphQLHandler
}

// Save{{clean (upper .ModuleName)}} resolver
func (m *mutationResolver) Save{{clean (upper .ModuleName)}}(ctx context.Context, input struct{ Data shareddomain.{{clean (upper .ModuleName)}} }) (ok string, err error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:Save{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.root.uc.Save{{clean (upper .ModuleName)}}(ctx, &input.Data); err != nil {
		return ok, err
	}
	return "Success", nil
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
		output <- "Hello from {{clean (upper .ModuleName)}}"
	}()

	return output
}
`

	deliveryGraphqlFieldResolverTemplate = `package graphqlhandler

import (
	shareddomain "{{.PackagePrefix}}/pkg/shared/domain"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
)

// CommonFilter basic filter model
type CommonFilter struct {
	Limit   *int
	Page    *int
	Search  *string
	Sort    *string
	ShowAll *bool
	OrderBy *string
}

// toSharedFilter method
func (f *CommonFilter) toSharedFilter() (filter candishared.Filter) {
	filter.Search = candihelper.PtrToString(f.Search)
	filter.OrderBy = candihelper.PtrToString(f.OrderBy)
	filter.Sort = candihelper.PtrToString(f.Sort)
	filter.ShowAll = candihelper.PtrToBool(f.ShowAll)

	if f.Limit == nil {
		filter.Limit = 10
	} else {
		filter.Limit = *f.Limit
	}
	if f.Page == nil {
		filter.Page = 1
	} else {
		filter.Page = *f.Page
	}

	return
}

// {{clean (upper .ModuleName)}}ListResolver resolver
type {{clean (upper .ModuleName)}}ListResolver struct {
	Meta candishared.Meta
	Data []shareddomain.{{clean (upper .ModuleName)}}
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
	{{clean $module.ModuleName}}: {{clean (upper $module.ModuleName)}}QueryResolver
{{- end }}
}

type Mutation {
{{- range $module := .Modules}}
	{{clean $module.ModuleName}}: {{clean (upper $module.ModuleName)}}MutationResolver
{{- end }}
}

type Subscription {
{{- range $module := .Modules}}
	{{clean $module.ModuleName}}: {{clean (upper $module.ModuleName)}}SubscriptionResolver
{{- end }}
}
`

	defaultGraphqlSchema = `# {{.Header}}

# {{clean (upper .ModuleName)}}Module Resolver Area
type {{clean (upper .ModuleName)}}QueryResolver {
	get_all_{{clean .ModuleName}}(filter: FilterListInputResolver): {{clean (upper .ModuleName)}}ListResolver!
	get_detail_{{clean .ModuleName}}(id: String!): {{clean (upper .ModuleName)}}Resolver!
}

type {{clean (upper .ModuleName)}}MutationResolver {
	save_{{clean .ModuleName}}(data: {{clean (upper .ModuleName)}}InputResolver!): String!
}

type {{clean (upper .ModuleName)}}SubscriptionResolver {
	hello(): String!
}

type {{clean (upper .ModuleName)}}ListResolver {
	meta: MetaResolver!
	data: [{{clean (upper .ModuleName)}}Resolver!]!
}

type {{clean (upper .ModuleName)}}Resolver {
	id: String!
}

input {{clean (upper .ModuleName)}}InputResolver {
	id: String!
}
`

	templateGraphqlCommon = `# {{.Header}}

input FilterListInputResolver {
	limit: Int
	page: Int
	"Optional (asc desc)"
	sort: FilterSortEnum
	"Optional"
	order_by: String
	"Optional"
	show_all: Boolean
	"Optional"
	search: String
}

type MetaResolver {
	page: Int!
	limit: Int!
	total_records: Int!
	total_pages: Int!
}

enum FilterSortEnum {
	asc
	desc
}
`
)
