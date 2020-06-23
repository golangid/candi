package interfaces

import (
	"context"

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

// WorkerHandler delivery factory for all worker handler
type WorkerHandler interface {
	GetTopics() []string
	ProcessMessage(ctx context.Context, topic string, message []byte)
}
