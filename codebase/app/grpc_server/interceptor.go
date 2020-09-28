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
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/helper"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

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
func unaryTracerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
		span = globalTracer.StartSpan(opName, ext.RPCServerOption((spanCtx)))
		ctx = opentracing.ContextWithSpan(ctx, span)
		ext.SpanKindRPCClient.Set(span)
	}
	defer span.Finish()

	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		span.SetTag("metadata", string(helper.ToBytes(meta)))
	}

	span.SetTag("req.body", req)

	resp, err = handler(ctx, req)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("error.value", err)
	}
	return
}

func unaryLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	defer func() {
		logInterceptor(start, err, info.FullMethod, "GRPC")
	}()

	resp, err = handler(ctx, req)
	return
}

func unaryPanicInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = grpc.Errorf(codes.Aborted, "%v", r)
		}
	}()

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
func streamTracerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	ctx := stream.Context()
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Aborted, "missing context metadata")
	}
	globalTracer := opentracing.GlobalTracer()
	opName := fmt.Sprintf("GRPC-Stream: %s", info.FullMethod)

	var span opentracing.Span
	if spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, tracer.GRPCMetadataReaderWriter(meta)); err != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, opName)
		ext.SpanKindRPCServer.Set(span)
	} else {
		span = globalTracer.StartSpan(opName, ext.RPCServerOption((spanCtx)))
		ctx = opentracing.ContextWithSpan(ctx, span)
		ext.SpanKindRPCClient.Set(span)
	}
	defer span.Finish()

	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		span.SetTag("metadata", string(helper.ToBytes(meta)))
	}

	err = handler(srv, stream)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("error.value", err)
	}
	return
}

func streamLogInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	start := time.Now()
	defer func() {
		logInterceptor(start, err, info.FullMethod, "GRPC-STREAM")
	}()

	return handler(srv, stream)
}

func streamPanicInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = grpc.Errorf(codes.Aborted, "%v", r)
		}
	}()

	return handler(srv, stream)
}

// Log incoming grpc request
func logInterceptor(startTime time.Time, err error, fullMethod string, reqType string) {
	end := time.Now()
	var status = "OK"
	statusColor := helper.Green
	if err != nil {
		statusColor = helper.Red
		status = "ERROR"
	}

	fmt.Fprintf(os.Stdout, "%s[%s]%s :%d %v | %s %-5s %s | %13v | %s\n",
		helper.Cyan, reqType, helper.Reset, config.BaseEnv().GRPCPort,
		end.Format("2006/01/02 - 15:04:05"),
		statusColor, status, helper.Reset,
		end.Sub(startTime),
		fullMethod,
	)
}
