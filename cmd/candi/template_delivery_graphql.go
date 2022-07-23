package main

const (
	deliveryGraphqlRootTemplate = `// {{.Header}}

package graphqlhandler

import (
	"{{.PackagePrefix}}/pkg/shared/usecase"
	
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
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

// RegisterMiddleware register resolver based on schema in "api/graphql/*" path
func (h *GraphQLHandler) RegisterMiddleware(mwGroup *types.MiddlewareGroup) {
	mwGroup.Add("{{upper (camel .ModuleName)}}QueryResolver.getAll{{upper (camel .ModuleName)}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{upper (camel .ModuleName)}}QueryResolver.getDetail{{upper (camel .ModuleName)}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{upper (camel .ModuleName)}}MutationResolver.create{{upper (camel .ModuleName)}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{upper (camel .ModuleName)}}MutationResolver.update{{upper (camel .ModuleName)}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{upper (camel .ModuleName)}}MutationResolver.delete{{upper (camel .ModuleName)}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
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

	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/tracer"
)

type queryResolver struct {
	root *GraphQLHandler
}

// GetAll{{upper (camel .ModuleName)}} resolver
func (q *queryResolver) GetAll{{upper (camel .ModuleName)}}(ctx context.Context, input struct{ Filter *CommonFilter }) (results {{upper (camel .ModuleName)}}ListResolver, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:GetAll{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if input.Filter == nil {
		input.Filter = new(CommonFilter)
	}
	filter := input.Filter.toSharedFilter()
	if err := q.root.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/get_all", filter); err != nil {
		return results, err
	}
	data, meta, err := q.root.uc.{{upper (camel .ModuleName)}}().GetAll{{upper (camel .ModuleName)}}(ctx, &filter)
	if err != nil {
		return results, err
	}

	return {{upper (camel .ModuleName)}}ListResolver{
		Meta: meta, Data: data,
	}, nil
}

// GetDetail{{upper (camel .ModuleName)}} resolver
func (q *queryResolver) GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, input struct{ ID string }) (data domain.Response{{upper (camel .ModuleName)}}, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:GetDetail{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	return q.root.uc.{{upper (camel .ModuleName)}}().GetDetail{{upper (camel .ModuleName)}}(ctx, input.ID)
}
`
	deliveryGraphqlMutationTemplate = `// {{.Header}}

package graphqlhandler

import (
	"context"
	
	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

	"{{.LibraryName}}/tracer"
)

type mutationResolver struct {
	root *GraphQLHandler
}

// Create{{upper (camel .ModuleName)}} resolver
func (m *mutationResolver) Create{{upper (camel .ModuleName)}}(ctx context.Context, input struct{ Data domain.Request{{upper (camel .ModuleName)}} }) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.root.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", input.Data); err != nil {
		return "", err
	}
	if err := m.root.uc.{{upper (camel .ModuleName)}}().Create{{upper (camel .ModuleName)}}(ctx, &input.Data); err != nil {
		return ok, err
	}
	return "Success", nil
}

// Update{{upper (camel .ModuleName)}} resolver
func (m *mutationResolver) Update{{upper (camel .ModuleName)}}(ctx context.Context, input struct {
	ID   string
	Data domain.Request{{upper (camel .ModuleName)}}
}) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	input.Data.ID = input.ID
	if err := m.root.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", input.Data); err != nil {
		return "", err
	}
	if err := m.root.uc.{{upper (camel .ModuleName)}}().Update{{upper (camel .ModuleName)}}(ctx, &input.Data); err != nil {
		return ok, err
	}
	return "Success", nil
}

// Delete{{upper (camel .ModuleName)}} resolver
func (m *mutationResolver) Delete{{upper (camel .ModuleName)}}(ctx context.Context, input struct{ ID string }) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGraphQL:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.root.uc.{{upper (camel .ModuleName)}}().Delete{{upper (camel .ModuleName)}}(ctx, input.ID); err != nil {
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

	shareddomain "{{.PackagePrefix}}/pkg/shared/domain"` + `
	
	{{if and .MongoDeps (not .SQLDeps)}}"go.mongodb.org/mongo-driver/bson/primitive"{{else}}"github.com/google/uuid"{{end}}` + `

	"{{.LibraryName}}/logger"
)

type subscriptionResolver struct {
	root *GraphQLHandler
}

// ListenData resolver, broadcast event to client
func (s *subscriptionResolver) ListenData(ctx context.Context) <-chan shareddomain.{{upper (camel .ModuleName)}} {
	output := make(chan shareddomain.{{upper (camel .ModuleName)}})

	go func() {
		// example send event to client every 5 seconds
		tick := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-tick.C:
				data := shareddomain.{{upper (camel .ModuleName)}}{
					CreatedAt:  time.Now(),
					UpdatedAt: time.Now(),
				}
				data.ID = {{if and .MongoDeps (not .SQLDeps)}}primitive.NewObjectID(){{else}}uuid.NewString(){{end}}
				output <- data
			case <-ctx.Done():
				logger.LogI("done")
				return
			}
		}
	}()

	return output
}
`

	deliveryGraphqlFieldResolverTemplate = `package graphqlhandler

import (
	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"

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
func (f *CommonFilter) toSharedFilter() (filter domain.Filter{{upper (camel .ModuleName)}}) {
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

// {{upper (camel .ModuleName)}}ListResolver resolver
type {{upper (camel .ModuleName)}}ListResolver struct {
	Meta candishared.Meta
	Data []domain.Response{{upper (camel .ModuleName)}}
}
`

	defaultGraphqlRootSchema = `# {{.Header}}

schema {
	query: Query
	mutation: Mutation
	subscription: Subscription
}

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
	getAll{{upper (camel .ModuleName)}}(filter: FilterListInputResolver): {{upper (camel .ModuleName)}}ListResolver!
	getDetail{{upper (camel .ModuleName)}}(id: String!): {{upper (camel .ModuleName)}}Resolver!
}

type {{upper (camel .ModuleName)}}MutationResolver {
	create{{upper (camel .ModuleName)}}(data: {{upper (camel .ModuleName)}}InputResolver!): String!
	update{{upper (camel .ModuleName)}}(id: String!, data: {{upper (camel .ModuleName)}}InputResolver!): String!
	delete{{upper (camel .ModuleName)}}(id: String!): String!
}

type {{upper (camel .ModuleName)}}SubscriptionResolver {
	listenData(): {{upper (camel .ModuleName)}}Resolver!
}

type {{upper (camel .ModuleName)}}ListResolver {
	meta: MetaResolver!
	data: [{{upper (camel .ModuleName)}}Resolver!]!
}

type {{upper (camel .ModuleName)}}Resolver {
	id: String!
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
