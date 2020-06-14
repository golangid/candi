package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// GraphQLHandler model
type GraphQLHandler struct {
	uc usecase.PublicUsecase
	mw middleware.Middleware
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw middleware.Middleware, uc usecase.PublicUsecase) *GraphQLHandler {
	return &GraphQLHandler{
		uc: uc,
		mw: mw,
	}
}

// GetHomePage handler
func (h *GraphQLHandler) GetHomePage(ctx context.Context) (*HomepageResolver, error) {
	homepage := h.uc.GetHomePage(ctx)
	return &HomepageResolver{
		Content: homepage.Content,
		Skills:  homepage.Skills,
		Footer:  homepage.Footer,
	}, nil
}

// GetAllVisitor handler
func (h *GraphQLHandler) GetAllVisitor(ctx context.Context, filter struct{ *shared.Filter }) (*VisitorListResolver, error) {
	h.mw.GraphQLBasicAuth(ctx)

	visitors, meta, err := h.uc.GetAllVisitor(ctx, filter.Filter)
	if err != nil {
		return nil, err
	}

	var visitorResolvers []*VisitorResolver
	for _, visitor := range visitors {
		visitorResolvers = append(visitorResolvers, &VisitorResolver{
			v: visitor,
		})
	}

	resolvers := VisitorListResolver{
		m:      meta,
		events: visitorResolvers,
	}
	return &resolvers, nil
}
