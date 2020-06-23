package grpchandler

import (
	"context"

	proto "agungdwiprasetyo.com/backend-microservices/api/notification-service/proto/push-notif"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"google.golang.org/grpc"
)

// GRPCHandler rpc handler
type GRPCHandler struct {
	mw interfaces.Middleware
}

// NewGRPCHandler func
func NewGRPCHandler(mw interfaces.Middleware) *GRPCHandler {
	return &GRPCHandler{
		mw: mw,
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server) {
	proto.RegisterPushNotifHandlerServer(server, h)
}

// FindAll rpc
func (h *GRPCHandler) Hello(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	return &proto.Response{
		Message: req.Message + "; Hello, from service: notification-service, module: push-notif",
	}, nil
}

