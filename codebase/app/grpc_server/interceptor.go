package grpcserver

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type interceptor struct {
	middleware types.MiddlewareGroup
	opt        *option
}

// for unary server
// chainUnaryServer creates a single interceptor out of a chain of many interceptors.
func chainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		chainer := func(currentInter grpc.UnaryServerInterceptor, currentHandler grpc.UnaryHandler) grpc.UnaryHandler {
			return func(currentCtx context.Context, currentReq any) (any, error) {
				return currentInter(currentCtx, currentReq, info, currentHandler)
			}
		}

		chainedHandler := handler
		for i := n - 1; i >= 0; i-- {
			chainedHandler = chainer(interceptors[i], chainedHandler)
		}

		return chainedHandler(ctx, req)
	}
}

// unaryTracerInterceptor for extract incoming tracer
func (i *interceptor) unaryTracerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	start := time.Now()
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	if metaDisableTrace := meta.Get(candihelper.HeaderDisableTrace); len(metaDisableTrace) > 0 {
		isDisableTrace, _ := strconv.ParseBool(metaDisableTrace[0])
		if isDisableTrace {
			ctx = candishared.SetToContext(tracer.SkipTraceContext(ctx), candishared.ContextKey(candihelper.HeaderDisableTrace), true)
			return handler(ctx, req)
		}
	}

	header := make(map[string]string, meta.Len())
	for key, values := range meta {
		for _, value := range values {
			header[key] = value
		}
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "GRPC-Server", header)
	defer func() {
		if rec := recover(); rec != nil {
			trace.SetTag("panic", true)
			err = status.Errorf(codes.Aborted, "%v", rec)
		}
		i.logInterceptor(start, err, info.FullMethod, "GRPC")
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("method", info.FullMethod)
	trace.Log("metadata", meta)
	if reqBody := candihelper.ToBytes(req); len(reqBody) < i.opt.jaegerMaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
		trace.Log("request.body", reqBody)
	} else {
		trace.Log("request.body.size", len(reqBody))
	}

	resp, err = handler(ctx, req)
	if respBody := candihelper.ToBytes(resp); len(respBody) < i.opt.jaegerMaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
		trace.Log("response.body", respBody)
	} else {
		trace.Log("response.body.size", len(respBody))
	}
	return
}

func (i *interceptor) unaryMiddlewareInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	ctx, err = i.middlewareInterceptor(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}

	resp, err = handler(ctx, req)
	return
}

// for stream server
// chainStreamServer creates a single interceptor out of a chain of many interceptors.
func chainStreamServer(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	n := len(interceptors)

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		chainer := func(currentInter grpc.StreamServerInterceptor, currentHandler grpc.StreamHandler) grpc.StreamHandler {
			return func(currentSrv any, currentStream grpc.ServerStream) error {
				return currentInter(currentSrv, currentStream, info, currentHandler)
			}
		}

		chainedHandler := handler
		for i := n - 1; i >= 0; i-- {
			chainedHandler = chainer(interceptors[i], chainedHandler)
		}

		return chainedHandler(srv, ss)
	}
}

// streamTracerInterceptor for extract incoming tracer
func (i *interceptor) streamTracerInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	start := time.Now()
	ctx := stream.Context()
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(srv, stream)
	}

	if metaDisableTrace := meta.Get(candihelper.HeaderDisableTrace); len(metaDisableTrace) > 0 {
		isDisableTrace, _ := strconv.ParseBool(metaDisableTrace[0])
		if isDisableTrace {
			return handler(srv, &wrappedServerStream{
				ServerStream: stream, wrappedContext: tracer.SkipTraceContext(ctx),
			})
		}
	}

	header := make(map[string]string, meta.Len())
	for key, values := range meta {
		for _, value := range values {
			header[key] = value
		}
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "GRPC-STREAM", header)
	defer func() {
		if rec := recover(); rec != nil {
			trace.SetTag("panic", true)
			err = status.Errorf(codes.Aborted, "%v", rec)
		}
		i.logInterceptor(start, err, info.FullMethod, "GRPC-STREAM")
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("method", info.FullMethod)
	trace.Log("metadata", meta)
	err = handler(srv, &wrappedServerStream{ServerStream: stream, wrappedContext: ctx})
	return
}

func (i *interceptor) streamMiddlewareInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	ctx, err := i.middlewareInterceptor(stream.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	return handler(srv, &wrappedServerStream{ServerStream: stream, wrappedContext: ctx})
}

func (i *interceptor) middlewareInterceptor(ctx context.Context, fullMethod string) (context.Context, error) {
	var err error

	if middFunc, ok := i.middleware[fullMethod]; ok {
		for _, mw := range middFunc {
			ctx, err = mw(ctx)
			if err != nil {
				return ctx, err
			}
		}
	}

	return ctx, nil
}

// Log incoming grpc request
func (i *interceptor) logInterceptor(startTime time.Time, err error, fullMethod string, reqType string) {
	if !i.opt.debugMode {
		return
	}

	end := time.Now()
	var status = "OK"
	statusColor := []byte{27, 91, 57, 55, 59, 52, 50, 109} // green
	if err != nil {
		statusColor = []byte{27, 91, 57, 55, 59, 52, 49, 109} // red
		status = "ERROR"
	}

	fmt.Fprintf(os.Stdout, "%s[%s]%s %s %v | %s %-5s %s | %13v | %s\n",
		[]byte{27, 91, 57, 55, 59, 52, 54, 109}, // cyan
		reqType,
		[]byte{27, 91, 48, 109}, // reset
		i.opt.tcpPort,
		end.Format("2006/01/02 - 15:04:05"),
		statusColor, status,
		[]byte{27, 91, 48, 109}, // reset
		end.Sub(startTime),
		fullMethod,
	)
}

// wrappedServerStream for inject custom context and wrap stream server
type wrappedServerStream struct {
	grpc.ServerStream
	wrappedContext context.Context
}

// Context get context
func (w *wrappedServerStream) Context() context.Context {
	return w.wrappedContext
}
