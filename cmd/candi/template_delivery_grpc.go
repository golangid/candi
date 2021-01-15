package main

const (
	deliveryGRPCTemplate = `// {{.Header}}

package grpchandler

import (
	"context"

	proto "{{.GoModName}}/api/proto/{{.ModuleName}}"
	"{{.GoModName}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"

	"google.golang.org/grpc"

	"{{.PackageName}}/candishared"
	"{{.PackageName}}/codebase/factory/types"
	"{{.PackageName}}/codebase/interfaces"
	"{{.PackageName}}/tracer"
)

// GRPCHandler rpc handler
type GRPCHandler struct {
	mw        interfaces.Middleware
	uc        usecase.{{clean (upper .ModuleName)}}Usecase
	validator interfaces.Validator
}

// NewGRPCHandler func
func NewGRPCHandler(mw interfaces.Middleware, uc usecase.{{clean (upper .ModuleName)}}Usecase, validator interfaces.Validator) *GRPCHandler {
	return &GRPCHandler{
		mw: mw, uc: uc, validator: validator,
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server, mwGroup *types.MiddlewareGroup) {
	proto.Register{{clean (upper .ModuleName)}}HandlerServer(server, h)

	// register middleware for method
	mwGroup.AddProto(proto.File_{{cleanPathModule .ModuleName}}_{{cleanPathModule .ModuleName}}_proto, "Hello", h.mw.GRPCBearerAuth)
}

// Hello rpc method
func (h *GRPCHandler) Hello(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	return &proto.Response{
		Message: h.uc.Hello(ctx) + ", with your session (" + tokenClaim.Audience + ")",
	}, nil
}
`

	defaultGRPCProto = `syntax="proto3";
package {{clean .ModuleName}};
option go_package = "{{.GoModName}}/api/proto/{{.ModuleName}}";

service {{clean (upper .ModuleName)}}Handler {
	rpc Hello(Request) returns (Response);
}

message Request {
    string Message=1;
}

message Response {
	string Message=1;
}`
)
