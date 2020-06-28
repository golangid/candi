package grpchandler

import (
	"context"

	proto "agungdwiprasetyo.com/backend-microservices/api/auth-service/proto/token"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"google.golang.org/grpc"
)

// GRPCHandler rpc handler
type GRPCHandler struct {
	mw interfaces.Middleware
	uc usecase.TokenUsecase
}

// NewGRPCHandler func
func NewGRPCHandler(mw interfaces.Middleware, uc usecase.TokenUsecase) *GRPCHandler {
	return &GRPCHandler{
		mw: mw,
		uc: uc,
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server) {
	proto.RegisterTokenHandlerServer(server, h)
}

// Hello rpc
func (h *GRPCHandler) Hello(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	return &proto.Response{
		Message: req.Message + "; Hello, from service: auth-service, module: token",
	}, nil
}

// ValidateToken rpc
func (h *GRPCHandler) ValidateToken(ctx context.Context, req *proto.PayloadValidate) (*proto.ResponseValidation, error) {

	result := <-h.uc.Validate(ctx, req.Token)
	if result.Error != nil {
		return nil, result.Error
	}

	return &proto.ResponseValidation{
		Success: true,
		Claim: &proto.ResponseValidation_ClaimData{
			Audience: "admin",
			Subject:  "user",
			User: &proto.ResponseValidation_ClaimData_UserData{
				ID: "001",
			},
		},
	}, nil
}
