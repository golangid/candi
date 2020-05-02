package graphqlhandler

import (
	"context"
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// GraphQLHandler model
type GraphQLHandler struct {
	uc        usecase.PublicUsecase
	basicAuth func(ctx context.Context)
}

// NewGraphQLHandler delivery
func NewGraphQLHandler(mw middleware.Middleware, uc usecase.PublicUsecase) *GraphQLHandler {
	return &GraphQLHandler{
		uc: uc,
		basicAuth: func(ctx context.Context) {
			headers := ctx.Value(shared.ContextKey("headers")).(http.Header)
			if err := mw.BasicAuth(headers.Get("Authorization")); err != nil {
				panic(err)
			}
		},
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
