package main

const (
	deliveryGraphqlRootTemplate = `// {{.Header}}

package graphqlhandler

import (
	"{{.PackagePrefix}}/pkg/shared/usecase"
	
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/interfaces"
)

// GraphQLHandler model
type GraphQLHandler struct {
	mw        interfaces.Middleware
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(uc usecase.Usecase, deps dependency.Dependency) *GraphQLHandler {
	return &GraphQLHandler{
		uc: uc, mw: deps.GetMiddleware(), validator: deps.GetValidator(),
	}
}

// Query method
func (h *GraphQLHandler) Query() interface{} {
	return h
}

// Mutation method
func (h *GraphQLHandler) Mutation() interface{} {
	return h
}

// Subscription method
func (h *GraphQLHandler) Subscription() interface{} {
	return h
}

// Schema method
func (h *GraphQLHandler) Schema() string {
	return ""
}
`

	deliveryGraphqlQueryTemplate = `// {{.Header}}

package graphqlhandler

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/tracer"
)

// GetAll{{upper (camel .ModuleName)}} resolver
// GetAll{{upper (camel .ModuleName)}} resolver
func (q *GraphQLHandler) GetAll{{upper (camel .ModuleName)}}(ctx context.Context, input struct {
	Filter *struct {
		candishared.NullableFilter
		domain.Filter{{upper (camel .ModuleName)}}
	}
}) (res struct {
	Meta candishared.Meta
	Data []domain.Response{{upper (camel .ModuleName)}}
}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:GetAll{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	filter := candihelper.UnwrapPtr(input.Filter)
	filter.Filter = filter.ToFilter()
	if err := q.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/get_all", filter.Filter{{upper (camel .ModuleName)}}); err != nil {
		return res, err
	}
	data, meta, err := q.uc.{{upper (camel .ModuleName)}}().GetAll{{upper (camel .ModuleName)}}(ctx, &filter.Filter{{upper (camel .ModuleName)}})
	if err != nil {
		return res, err
	}

	res.Data = data
	res.Meta = meta
	return
}

// GetDetail{{upper (camel .ModuleName)}} resolver
func (q *GraphQLHandler) GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, input struct{ ID {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}} }) (data domain.Response{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:GetDetail{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	return q.uc.{{upper (camel .ModuleName)}}().GetDetail{{upper (camel .ModuleName)}}(ctx, input.ID)
}
`
	deliveryGraphqlMutationTemplate = `// {{.Header}}

package graphqlhandler

import (
	"context"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/tracer"
)

// Create{{upper (camel .ModuleName)}} resolver
func (m *GraphQLHandler) Create{{upper (camel .ModuleName)}}(ctx context.Context, input struct{ Data domain.Request{{upper (camel .ModuleName)}} }) (data domain.Response{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", input.Data); err != nil {
		return data, err
	}
	return m.uc.{{upper (camel .ModuleName)}}().Create{{upper (camel .ModuleName)}}(ctx, &input.Data)
}

// Update{{upper (camel .ModuleName)}} resolver
func (m *GraphQLHandler) Update{{upper (camel .ModuleName)}}(ctx context.Context, input struct {
	ID   {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}}
	Data domain.Request{{upper (camel .ModuleName)}}
}) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	input.Data.ID = input.ID
	if err := m.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", input.Data); err != nil {
		return "", err
	}
	if err := m.uc.{{upper (camel .ModuleName)}}().Update{{upper (camel .ModuleName)}}(ctx, &input.Data); err != nil {
		return ok, err
	}
	return "Success", nil
}

// Delete{{upper (camel .ModuleName)}} resolver
func (m *GraphQLHandler) Delete{{upper (camel .ModuleName)}}(ctx context.Context, input struct{ ID {{if and .MongoDeps (not .SQLDeps)}}string{{else}}int{{end}} }) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.uc.{{upper (camel .ModuleName)}}().Delete{{upper (camel .ModuleName)}}(ctx, input.ID); err != nil {
		return ok, err
	}
	return "Success", nil
}
`
	deliveryGraphqlSubscriptionTemplate = `// {{.Header}}

package graphqlhandler

import (
	"context"
	"time"

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/logger"{{if and .MongoDeps (not .SQLDeps)}}

	"github.com/google/uuid"{{end}}
)

// ListenData resolver, broadcast event to client
func (s *GraphQLHandler) ListenData(ctx context.Context) <-chan domain.Response{{upper (camel .ModuleName)}} {
	output := make(chan domain.Response{{upper (camel .ModuleName)}})

	go func() {
		// example send event to client every 5 seconds
		tick := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-tick.C:
				data := domain.Response{{upper (camel .ModuleName)}}{
					CreatedAt: time.Now().Format(time.RFC3339),
					UpdatedAt: time.Now().Format(time.RFC3339),
				}
				data.ID = {{if and .MongoDeps (not .SQLDeps)}}uuid.NewString(){{else}}1{{end}}
				output <- data
			case <-ctx.Done():
				tick.Stop()
				logger.LogI("done")
				return
			}
		}
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

enum AuthTypeDirective {
	BASIC
	BEARER
	MULTIPLE
}
directive @auth(authType: AuthTypeDirective!) on FIELD_DEFINITION
directive @permissionACL(permissionCode: String!) on FIELD_DEFINITION

type Query {
	# @candi:queryRoot
}

type Mutation {
	# @candi:mutationRoot
}

type Subscription {
	# @candi:subscriptionRoot
}
`

	defaultGraphqlSchema = `# {{.Header}}

# {{upper (camel .ModuleName)}}Module Resolver Area
type {{upper (camel .ModuleName)}}QueryResolver {
	getAll{{upper (camel .ModuleName)}}(filter: FilterListInputResolver): {{upper (camel .ModuleName)}}ListResolver! @permissionACL(permissionCode: getAll{{upper (camel .ModuleName)}})
	getDetail{{upper (camel .ModuleName)}}(id: {{if and .MongoDeps (not .SQLDeps)}}String{{else}}Int{{end}}!): {{upper (camel .ModuleName)}}Resolver! @permissionACL(permissionCode: getDetail{{upper (camel .ModuleName)}})
}

type {{upper (camel .ModuleName)}}MutationResolver {
	create{{upper (camel .ModuleName)}}(data: {{upper (camel .ModuleName)}}InputResolver!): {{upper (camel .ModuleName)}}Resolver! @permissionACL(permissionCode: create{{upper (camel .ModuleName)}})
	update{{upper (camel .ModuleName)}}(id: {{if and .MongoDeps (not .SQLDeps)}}String{{else}}Int{{end}}!, data: {{upper (camel .ModuleName)}}InputResolver!): String! @permissionACL(permissionCode: update{{upper (camel .ModuleName)}})
	delete{{upper (camel .ModuleName)}}(id: {{if and .MongoDeps (not .SQLDeps)}}String{{else}}Int{{end}}!): String! @permissionACL(permissionCode: delete{{upper (camel .ModuleName)}})
}

type {{upper (camel .ModuleName)}}SubscriptionResolver {
	listenData(): {{upper (camel .ModuleName)}}Resolver!
}

type {{upper (camel .ModuleName)}}ListResolver {
	meta: MetaResolver!
	data: [{{upper (camel .ModuleName)}}Resolver!]!
}

type {{upper (camel .ModuleName)}}Resolver {
	id: {{if and .MongoDeps (not .SQLDeps)}}String{{else}}Int{{end}}!
	field: String!
	createdAt: String!
	updatedAt: String!
}

input {{upper (camel .ModuleName)}}InputResolver {
	field: String!
}
`

	templateGraphqlCommon = `# {{.Header}}

input FilterListInputResolver {
	limit: Int
	page: Int
	"Optional (asc desc)"
	sort: FilterSortEnum
	"Optional"
	orderBy: String
	"Optional"
	showAll: Boolean
	"Optional"
	search: String
}

type MetaResolver {
	page: Int!
	limit: Int!
	totalRecords: Int!
	totalPages: Int!
}

enum FilterSortEnum {
	ASC
	DESC
}
`
)
