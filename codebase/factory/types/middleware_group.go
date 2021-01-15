package types

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// MiddlewareFunc type
type MiddlewareFunc func(context.Context) context.Context

// MiddlewareGroup type
type MiddlewareGroup map[string]MiddlewareFunc

// Add register full method to middleware
func (mw MiddlewareGroup) Add(fullMethod string, middlewareFunc MiddlewareFunc) {
	mw[fullMethod] = middlewareFunc
}

// AddProto register proto for grpc middleware
func (mw MiddlewareGroup) AddProto(protoDesc protoreflect.FileDescriptor, method string, middlewareFunc MiddlewareFunc) {
	serviceName := fmt.Sprintf("/%s/", protoDesc.Services().Get(0).FullName())
	mw[serviceName+method] = middlewareFunc
}
