package grpchandler

import (
	"context"

	proto "agungdwiprasetyo.com/backend-microservices/api/user-service/proto/customer"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"google.golang.org/grpc"
)

// GRPCHandler rpc handler
type GRPCHandler struct {
	mw middleware.Middleware
}

// NewGRPCHandler func
func NewGRPCHandler(mw middleware.Middleware) *GRPCHandler {
	return &GRPCHandler{
		mw: mw,
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server) {
	proto.RegisterCustomerHandlerServer(server, h)
}

// FindAll rpc
func (h *GRPCHandler) Hello(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	return &proto.Response{
		Message: req.Message + "; Hello, from service: user-service, module: customer",
	}, nil
}

