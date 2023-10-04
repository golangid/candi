package interfaces

import (
	"github.com/golangid/candi/codebase/factory/types"
	"google.golang.org/grpc"
)

// RESTHandler delivery factory for REST handler (default using echo rest framework)
type RESTHandler interface {
	Mount(group RESTRouter)
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
	Schema() string
}

// WorkerHandler delivery factory for all worker handler
type WorkerHandler interface {
	MountHandlers(group *types.WorkerHandlerGroup)
}

// ServerHandler delivery factory for all additional server handler (rest framework, p2p, and many more)
type ServerHandler interface {
	MountHandlers(group interface{}) // why interface? cause every server is different type for grouping route handler
}
