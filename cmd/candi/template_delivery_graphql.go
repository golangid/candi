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
	mwGroup.Add("{{clean (upper .ModuleName)}}MutationResolver.create_{{clean .ModuleName}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{clean (upper .ModuleName)}}MutationResolver.update_{{clean .ModuleName}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
	mwGroup.Add("{{clean (upper .ModuleName)}}MutationResolver.delete_{{clean .ModuleName}}", h.mw.GraphQLBearerAuth, h.mw.GraphQLPermissionACL("resource.public"))
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
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()

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
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:GetDetail{{clean (upper .ModuleName)}}")
	defer trace.Finish()

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

// Create{{clean (upper .ModuleName)}} resolver
func (m *mutationResolver) Create{{clean (upper .ModuleName)}}(ctx context.Context, input struct{ Data shareddomain.{{clean (upper .ModuleName)}} }) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:Create{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.root.uc.Create{{clean (upper .ModuleName)}}(ctx, &input.Data); err != nil {
		return ok, err
	}
	return "Success", nil
}

// Update{{clean (upper .ModuleName)}} resolver
func (m *mutationResolver) Update{{clean (upper .ModuleName)}}(ctx context.Context, input struct {
	ID   string
	Data shareddomain.{{clean (upper .ModuleName)}}
}) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:Update{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := golibshared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.root.uc.Update{{clean (upper .ModuleName)}}(ctx, input.ID, &input.Data); err != nil {
		return ok, err
	}
	return "Success", nil
}

// Delete{{clean (upper .ModuleName)}} resolver
func (m *mutationResolver) Delete{{clean (upper .ModuleName)}}(ctx context.Context, input struct{ ID string }) (ok string, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGraphQL:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GraphQLBearerAuth in middleware for this resolver

	if err := m.root.uc.Delete{{clean (upper .ModuleName)}}(ctx, input.ID); err != nil {
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

	shareddomain "{{.PackagePrefix}}/pkg/shared/domain"

	"github.com/google/uuid"
	"{{.LibraryName}}/logger"
)

type subscriptionResolver struct {
	root *GraphQLHandler
}

// ListenData resolver, broadcast event to client
func (s *subscriptionResolver) ListenData(ctx context.Context) <-chan shareddomain.{{clean (upper .ModuleName)}} {
	output := make(chan shareddomain.{{clean (upper .ModuleName)}})

	go func() {
		// example send event to client every 5 seconds
		tick := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-tick.C:
				output <- shareddomain.{{clean (upper .ModuleName)}}{
					ID:         uuid.NewString(),
					CreatedAt:  time.Now(),
					ModifiedAt: time.Now(),
				}
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

# {{clean (upper .ModuleName)}}Module Resolver Area
type {{clean (upper .ModuleName)}}QueryResolver {
	get_all_{{clean .ModuleName}}(filter: FilterListInputResolver): {{clean (upper .ModuleName)}}ListResolver!
	get_detail_{{clean .ModuleName}}(id: String!): {{clean (upper .ModuleName)}}Resolver!
}

type {{clean (upper .ModuleName)}}MutationResolver {
	create_{{clean .ModuleName}}(data: {{clean (upper .ModuleName)}}InputResolver!): String!
	update_{{clean .ModuleName}}(id: String!, data: {{clean (upper .ModuleName)}}InputResolver!): String!
	delete_{{clean .ModuleName}}(id: String!): String!
}

type {{clean (upper .ModuleName)}}SubscriptionResolver {
	listen_data(): {{clean (upper .ModuleName)}}Resolver!
}

type {{clean (upper .ModuleName)}}ListResolver {
	meta: MetaResolver!
	data: [{{clean (upper .ModuleName)}}Resolver!]!
}

type {{clean (upper .ModuleName)}}Resolver {
	id: String!
	created_at: String!
	modified_at: String!
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
