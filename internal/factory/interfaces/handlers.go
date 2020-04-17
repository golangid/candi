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

// SubscriberHandler delivery factory for all subscriber handler
type SubscriberHandler interface {
	GetTopics() []string
	ProcessMessage(ctx context.Context, message []byte)
}
