package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/golangid/candi/candihelper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	// BEARER constanta
	BEARER = "BEARER"

	// BASIC constanta
	BASIC = "BASIC"

	// MULTIPLE constanta
	MULTIPLE = "MULTIPLE"
)

func extractAuthType(prefix, authorization string) (string, error) {

	authValues := strings.Split(authorization, " ")
	if len(authValues) == 2 && strings.ToUpper(authValues[0]) == prefix {
		return authValues[1], nil
	}

	return "", errors.New("Invalid authorization")
}

func extractAuthorizationGRPCMetadata(ctx context.Context) (auth string, err error) {

	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return auth, grpc.Errorf(codes.Unauthenticated, "Missing context metadata")
	}

	authorizationMap := meta[strings.ToLower(candihelper.HeaderAuthorization)]
	if len(authorizationMap) != 1 {
		return auth, grpc.Errorf(codes.Unauthenticated, "Invalid authorization")
	}

	return authorizationMap[0], nil
}
