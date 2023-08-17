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
	"github.com/golangid/candi/logger"
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

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chainer := func(currentInter grpc.UnaryServerInterceptor, currentHandler grpc.UnaryHandler) grpc.UnaryHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
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
func (i *interceptor) unaryTracerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Aborted, "missing context metadata")
	}

	if metaDisableTrace := meta.Get(candihelper.HeaderDisableTrace); len(metaDisableTrace) > 0 {
		isDisableTrace, _ := strconv.ParseBool(metaDisableTrace[0])
		if isDisableTrace {
			ctx = candishared.SetToContext(tracer.SkipTraceContext(ctx), candishared.ContextKey(candihelper.HeaderDisableTrace), true)
			return handler(ctx, req)
		}
	}

	header := map[string]string{}
	for key, values := range meta {
		for _, value := range values {
			header[key] = value
		}
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, fmt.Sprintf("GRPC: %s", info.FullMethod), header)
	defer func() {
		if r := recover(); r != nil {
			trace.SetTag("panic", true)
			err = status.Errorf(codes.Aborted, "%v", r)
		}
		i.logInterceptor(start, err, info.FullMethod, "GRPC")
		logger.LogGreen("grpc > trace_url: " + tracer.GetTraceURL(ctx))
		if respBody := candihelper.ToBytes(resp); len(respBody) < i.opt.jaegerMaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
			trace.Log("response.body", respBody)
		} else {
			trace.Log("response.body.size", len(respBody))
		}
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("metadata", meta)
	if reqBody := candihelper.ToBytes(req); len(reqBody) < i.opt.jaegerMaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
		trace.Log("request.body", reqBody)
	} else {
		trace.Log("request.body.size", len(reqBody))
	}

	resp, err = handler(ctx, req)
	return
}

func (i *interceptor) unaryMiddlewareInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		chainer := func(currentInter grpc.StreamServerInterceptor, currentHandler grpc.StreamHandler) grpc.StreamHandler {
			return func(currentSrv interface{}, currentStream grpc.ServerStream) error {
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
func (i *interceptor) streamTracerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	start := time.Now()
	ctx := stream.Context()
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Aborted, "missing context metadata")
	}

	if metaDisableTrace := meta.Get(candihelper.HeaderDisableTrace); len(metaDisableTrace) > 0 {
		isDisableTrace, _ := strconv.ParseBool(metaDisableTrace[0])
		if isDisableTrace {
			return handler(srv, &wrappedServerStream{
				ServerStream: stream, wrappedContext: tracer.SkipTraceContext(ctx),
			})
		}
	}

	header := map[string]string{}
	for key, values := range meta {
		for _, value := range values {
			header[key] = value
		}
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, fmt.Sprintf("GRPC-STREAM: %s", info.FullMethod), header)
	defer func() {
		if r := recover(); r != nil {
			trace.SetTag("panic", true)
			err = status.Errorf(codes.Aborted, "%v", r)
		}
		i.logInterceptor(start, err, info.FullMethod, "GRPC-STREAM")
		logger.LogGreen("grpc_stream > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("metadata", meta)
	err = handler(srv, &wrappedServerStream{ServerStream: stream, wrappedContext: ctx})
	return
}

func (i *interceptor) streamMiddlewareInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	ctx, err := i.middlewareInterceptor(stream.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	return handler(srv, &wrappedServerStream{ServerStream: stream, wrappedContext: ctx})
}

func (i *interceptor) middlewareInterceptor(ctx context.Context, fullMethod string) (context.Context, error) {
	if middFunc, ok := i.middleware[fullMethod]; ok {
		for _, mw := range middFunc {
			ctx, err := mw(ctx)
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
	statusColor := candihelper.Green
	if err != nil {
		statusColor = candihelper.Red
		status = "ERROR"
	}

	fmt.Fprintf(os.Stdout, "%s[%s]%s %s %v | %s %-5s %s | %13v | %s\n",
		candihelper.Cyan, reqType, candihelper.Reset, i.opt.tcpPort,
		end.Format("2006/01/02 - 15:04:05"),
		statusColor, status, candihelper.Reset,
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
