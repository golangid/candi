package interfaces

import (
	"github.com/labstack/echo"
	"google.golang.org/grpc"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
)

// EchoRestHandler delivery factory for echo handler
type EchoRestHandler interface {
	Mount(group *echo.Group)
}

// GRPCHandler delivery factory for grpc handler
type GRPCHandler interface {
	Register(server *grpc.Server)
}

// GraphQLHandler delivery factory for graphql resolver handler
type GraphQLHandler interface {
	Query() interface{}
	Mutation() interface{}
	Subscription() interface{}
	RegisterMiddleware(group *types.GraphQLMiddlewareGroup)
}

// WorkerHandler delivery factory for all worker handler
type WorkerHandler interface {
	MountHandlers(group *types.WorkerHandlerGroup)
}
