package interfaces

import (
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
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
	RootName() string
	// waiting https://github.com/graph-gophers/graphql-go/issues/145 if include subscription in schema
	Query() interface{}
	Mutation() interface{}
	Subscription() interface{}
}

// WorkerHandler delivery factory for all worker handler
type WorkerHandler interface {
	MountHandlers() map[string]types.WorkerHandlerFunc
}
