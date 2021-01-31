package interfaces

import (
	"github.com/labstack/echo"
	"google.golang.org/grpc"
	"pkg.agungdp.dev/candi/codebase/factory/types"
)

// RESTHandler delivery factory for REST handler
type RESTHandler interface {
	Mount(group *echo.Group)
}

// GRPCHandler delivery factory for GRPC handler
type GRPCHandler interface {
	Register(server *grpc.Server, middlewareGroup *types.MiddlewareGroup)
}

// GraphQLHandler delivery factory for GraphQL resolver handler
type GraphQLHandler interface {
	Query() interface{}
	Mutation() interface{}
	Subscription() interface{}
	RegisterMiddleware(group *types.MiddlewareGroup)
}

// WorkerHandler delivery factory for all worker handler
type WorkerHandler interface {
	MountHandlers(group *types.WorkerHandlerGroup)
}
