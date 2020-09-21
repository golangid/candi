package auth

import (
	"context"
	"errors"

	pb "agungdwiprasetyo.com/backend-microservices/api/auth-service/proto/token"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authServiceGRPC struct {
	client  pb.TokenHandlerClient
	authKey string
}

// NewAuthServiceGRPC using redis
func NewAuthServiceGRPC(authGRPCHost, authServiceKey string) Auth {

	conn, err := grpc.Dial(authGRPCHost, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return &authServiceGRPC{
		client:  pb.NewTokenHandlerClient(conn),
		authKey: authServiceKey,
	}
}

func (a *authServiceGRPC) ValidateToken(ctx context.Context, token string) (cl *shared.TokenClaim, err error) {

	tracer.WithTraceFuncTracer(ctx, "AuthServiceGRPC-Validate", func(trace tracer.Tracer) {
		ctx = trace.Context()

		md := metadata.Pairs("authorization", a.authKey)
		trace.InjectGRPCMetadata(md)

		ctx = metadata.NewOutgoingContext(ctx, md)
		resp, err := a.client.ValidateToken(ctx, &pb.PayloadValidate{
			Token: token,
		})
		if err != nil {
			logger.LogE(err.Error())
			desc, ok := status.FromError(err)
			if ok {
				err = errors.New(desc.Message())
			}
			return
		}

		var sharedClaim shared.TokenClaim
		sharedClaim.Audience = resp.Claim.Audience
		sharedClaim.ExpiresAt = resp.Claim.ExpiresAt
		sharedClaim.IssuedAt = resp.Claim.IssuedAt
		sharedClaim.Issuer = resp.Claim.Issuer
		sharedClaim.NotBefore = resp.Claim.NotBefore
		sharedClaim.Subject = resp.Claim.Subject
		sharedClaim.User.ID = resp.Claim.User.ID
		sharedClaim.User.Username = resp.Claim.User.Username

		cl = &sharedClaim
	})

	return
}

func (a *authServiceGRPC) GenerateToken(ctx context.Context, claim *shared.TokenClaim) <-chan shared.Result {
	output := make(chan shared.Result)

	go tracer.WithTraceFuncTracer(ctx, "AuthServiceGRPC-Validate", func(trace tracer.Tracer) {
		defer close(output)
		ctx = trace.Context()

		md := metadata.Pairs("authorization", a.authKey)
		trace.InjectGRPCMetadata(md)

		ctx = metadata.NewOutgoingContext(ctx, md)
		resp, err := a.client.GenerateToken(ctx, &pb.UserData{
			ID:       claim.User.ID,
			Username: claim.User.Username,
		})
		if err != nil {
			logger.LogE(err.Error())
			desc, ok := status.FromError(err)
			if ok {
				err = errors.New(desc.Message())
			}
			output <- shared.Result{Error: err}
			return
		}

		output <- shared.Result{
			Data: resp.Data.Token,
		}
	})

	return output
}
