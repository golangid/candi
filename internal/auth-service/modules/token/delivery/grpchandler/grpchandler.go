package grpchandler

import (
	"context"

	proto "agungdwiprasetyo.com/backend-microservices/api/auth-service/proto/token"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

// GenerateToken rpc
func (h *GRPCHandler) GenerateToken(ctx context.Context, req *proto.UserData) (*proto.ResponseGenerate, error) {
	var tokenClaim shared.TokenClaim
	tokenClaim.User.ID = req.ID
	tokenClaim.User.Username = req.Username

	result := <-h.uc.Generate(ctx, &tokenClaim)
	if result.Error != nil {
		return nil, grpc.Errorf(codes.Internal, "%v", result.Error)
	}

	tokenString := result.Data.(string)

	return &proto.ResponseGenerate{
		Success: true,
		Data: &proto.ResponseGenerate_Token{
			Token: tokenString,
		},
	}, nil
}

// ValidateToken rpc
func (h *GRPCHandler) ValidateToken(ctx context.Context, req *proto.PayloadValidate) (*proto.ResponseValidation, error) {

	result := <-h.uc.Validate(ctx, req.Token)
	if result.Error != nil {
		return nil, result.Error
	}

	claim := result.Data.(*domain.Claim)

	return &proto.ResponseValidation{
		Success: true,
		Claim: &proto.ResponseValidation_ClaimData{
			Audience:  claim.Audience,
			Subject:   claim.Subject,
			ExpiresAt: claim.ExpiresAt,
			User: &proto.UserData{
				ID:       claim.User.ID,
				Username: claim.User.Username,
			},
		},
	}, nil
}
