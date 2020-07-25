package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type recvWrapper struct {
	ctx context.Context
	grpc.ServerStream
}

func (r *recvWrapper) Context() context.Context {
	return r.ctx
}

func TestMiddleware_GRPCAuth(t *testing.T) {

	t.Run("Testcase #1: Positive", func(t *testing.T) {
		md := metadata.Pairs("authorization", "dXNlcjpkYTFjMjVkOC0zN2M4LTQxYjEtYWZlMi00MmRkNDgyNWJmZWE=")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		unaryInfo := &grpc.UnaryServerInfo{
			FullMethod: "TestService.UnaryMethod",
		}
		unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "user-service", nil
		}

		mw := &Middleware{
			username: "user", password: "da1c25d8-37c8-41b1-afe2-42dd4825bfea",
		}
		_, err := mw.GRPCBasicAuth(ctx, "testing", unaryInfo, unaryHandler)
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative", func(t *testing.T) {
		md := metadata.Pairs("authorization", "invalid")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		unaryInfo := &grpc.UnaryServerInfo{
			FullMethod: "TestService.UnaryMethod",
		}
		unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "user-service", nil
		}

		mw := &Middleware{
			username: "user", password: "da1c25d8-37c8-41b1-afe2-42dd4825bfea",
		}
		_, err := mw.GRPCBasicAuth(ctx, "testing", unaryInfo, unaryHandler)
		assert.Error(t, err)
	})
}

func TestMiddleware_GRPCAuthStream(t *testing.T) {

	t.Run("Testcase #1: Positive", func(t *testing.T) {
		md := metadata.Pairs("authorization", "dXNlcjpkYTFjMjVkOC0zN2M4LTQxYjEtYWZlMi00MmRkNDgyNWJmZWE=")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		stream := &recvWrapper{
			ctx: ctx,
		}
		streamInfo := &grpc.StreamServerInfo{
			FullMethod: "TestService.StreamMethod",
		}
		streamHandler := func(srv interface{}, req grpc.ServerStream) error {
			return nil
		}

		mw := &Middleware{
			username: "user", password: "da1c25d8-37c8-41b1-afe2-42dd4825bfea",
		}
		err := mw.GRPCBasicAuthStream("test", stream, streamInfo, streamHandler)
		assert.NoError(t, err)
	})

	t.Run("Testcase #2: Negative, missing auth context", func(t *testing.T) {
		ctx := context.Background()

		stream := &recvWrapper{
			ctx: ctx,
		}
		streamInfo := &grpc.StreamServerInfo{
			FullMethod: "TestService.StreamMethod",
		}
		streamHandler := func(srv interface{}, req grpc.ServerStream) error {
			return nil
		}

		mw := &Middleware{
			username: "user", password: "da1c25d8-37c8-41b1-afe2-42dd4825bfea",
		}
		err := mw.GRPCBasicAuthStream("test", stream, streamInfo, streamHandler)
		assert.Error(t, err)
	})

	t.Run("Testcase #3: Negative, multiple authorization keys", func(t *testing.T) {
		md := metadata.Pairs("authorization", "dXNlcjpkYTFjMjVkOC0zN2M4LTQxYjEtYWZlMi00MmRkNDgyNWJmZWE=", "authorization", "double")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		stream := &recvWrapper{
			ctx: ctx,
		}
		streamInfo := &grpc.StreamServerInfo{
			FullMethod: "TestService.StreamMethod",
		}
		streamHandler := func(srv interface{}, req grpc.ServerStream) error {
			return nil
		}

		mw := &Middleware{
			username: "user", password: "da1c25d8-37c8-41b1-afe2-42dd4825bfea",
		}
		err := mw.GRPCBasicAuthStream("test", stream, streamInfo, streamHandler)
		assert.Error(t, err)
	})
}
