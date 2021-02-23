package grpcserver

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

type interceptor struct {
	middleware types.MiddlewareGroup
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
		return nil, grpc.Errorf(codes.Aborted, "missing context metadata")
	}
	globalTracer := opentracing.GlobalTracer()
	opName := fmt.Sprintf("GRPC: %s", info.FullMethod)

	var span opentracing.Span
	if spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, tracer.GRPCMetadataReaderWriter(meta)); err != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, opName)
		ext.SpanKindRPCServer.Set(span)
	} else {
		span = globalTracer.StartSpan(opName, opentracing.ChildOf(spanCtx), ext.SpanKindRPCClient)
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	defer func() {
		if r := recover(); r != nil {
			err = grpc.Errorf(codes.Aborted, "%v", r)
			tracer.SetError(ctx, err)
		}
		logInterceptor(start, err, info.FullMethod, "GRPC")
		logger.LogGreen("grpc > trace_url: " + tracer.GetTraceURL(ctx))
		span.LogKV("response.body", string(candihelper.ToBytes(resp)))
		span.Finish()
	}()

	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		span.SetTag("metadata", string(candihelper.ToBytes(meta)))
	}

	span.LogKV("request.body", string(candihelper.ToBytes(req)))

	resp, err = handler(ctx, req)
	if err != nil {
		tracer.SetError(ctx, err)
	}
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
		return grpc.Errorf(codes.Aborted, "missing context metadata")
	}
	globalTracer := opentracing.GlobalTracer()
	opName := fmt.Sprintf("GRPC-STREAM: %s", info.FullMethod)

	var span opentracing.Span
	if spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, tracer.GRPCMetadataReaderWriter(meta)); err != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, opName)
		ext.SpanKindRPCServer.Set(span)
	} else {
		span = globalTracer.StartSpan(opName, opentracing.ChildOf(spanCtx), ext.SpanKindRPCClient)
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	defer func() {
		if r := recover(); r != nil {
			err = grpc.Errorf(codes.Aborted, "%v", r)
			tracer.SetError(ctx, err)
		}
		logInterceptor(start, err, info.FullMethod, "GRPC-STREAM")
		logger.LogGreen("grpc_stream > trace_url: " + tracer.GetTraceURL(ctx))
		span.Finish()
	}()

	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		span.SetTag("metadata", string(candihelper.ToBytes(meta)))
	}

	err = handler(srv, &wrappedServerStream{ServerStream: stream, wrappedContext: ctx})
	if err != nil {
		tracer.SetError(ctx, err)
	}
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
		execMiddleware := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = grpc.Errorf(codes.Unauthenticated, "%v", r)
				}
			}()
			ctx = middFunc(ctx)
			return nil
		}
		if err := execMiddleware(); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

// Log incoming grpc request
func logInterceptor(startTime time.Time, err error, fullMethod string, reqType string) {
	if !env.BaseEnv().DebugMode {
		return
	}

	end := time.Now()
	var status = "OK"
	statusColor := candihelper.Green
	if err != nil {
		statusColor = candihelper.Red
		status = "ERROR"
	}

	fmt.Fprintf(os.Stdout, "%s[%s]%s :%d %v | %s %-5s %s | %13v | %s\n",
		candihelper.Cyan, reqType, candihelper.Reset, env.BaseEnv().GRPCPort,
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
