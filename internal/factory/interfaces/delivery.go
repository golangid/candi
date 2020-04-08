package interfaces

import (
	"context"

	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

// EchoRestDelivery delivery factory for echo handler
type EchoRestDelivery interface {
	Mount(group *echo.Group)
}

// GRPCDelivery delivery factory for grpc handler
type GRPCDelivery interface {
	Register(server *grpc.Server)
}

// SubscriberDelivery delivery factory for all subscriber handler
type SubscriberDelivery interface {
	GetTopics() []string
	ProcessMessage(ctx context.Context, message []byte)
}
