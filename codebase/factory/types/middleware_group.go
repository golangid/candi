package types

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"pkg.agungdp.dev/candi/candihelper"
)

// MiddlewareFunc type
type MiddlewareFunc func(context.Context) context.Context

// MiddlewareGroup type
type MiddlewareGroup map[string][]MiddlewareFunc

// Add register full method to middleware
func (mw MiddlewareGroup) Add(fullMethod string, middlewareFunc ...MiddlewareFunc) {
	mw[fullMethod] = middlewareFunc
}

// AddProto register proto for grpc middleware
func (mw MiddlewareGroup) AddProto(protoDesc protoreflect.FileDescriptor, handler interface{}, middlewareFunc ...MiddlewareFunc) {
	serviceName := fmt.Sprintf("/%s/", protoDesc.Services().Get(0).FullName())
	var fullMethodName string
	switch h := handler.(type) {
	case string:
		fullMethodName = serviceName + h
	default:
		fullMethodName = serviceName + candihelper.GetFuncName(handler)
	}
	mw[fullMethodName] = middlewareFunc
}
