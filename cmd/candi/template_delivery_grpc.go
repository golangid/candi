package main

const (
	deliveryGRPCTemplate = `// {{.Header}}

package grpchandler

import (
	"context"

	proto "{{.GoModName}}/api/proto/{{.ModuleName}}"
	"{{.GoModName}}/internal/modules/{{clean .ModuleName}}/usecase"

	"google.golang.org/grpc"

	"{{.PackageName}}/codebase/interfaces"
	"{{.PackageName}}/tracer"
)

// GRPCHandler rpc handler
type GRPCHandler struct {
	mw interfaces.Middleware
	uc usecase.{{clean (upper .ModuleName)}}Usecase
}

// NewGRPCHandler func
func NewGRPCHandler(mw interfaces.Middleware, uc usecase.{{clean (upper .ModuleName)}}Usecase) *GRPCHandler {
	return &GRPCHandler{
		mw: mw, uc: uc,
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server) {
	proto.Register{{clean (upper .ModuleName)}}HandlerServer(server, h)
}

// Hello rpc
func (h *GRPCHandler) Hello(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	trace := tracer.StartTrace(ctx, "DeliveryGRPC:Hello")
	defer trace.Finish()
	ctx = trace.Context()

	return &proto.Response{
		Message: req.Message + "; "+ h.uc.Hello(ctx),
	}, nil
}

`

	defaultGRPCProto = `syntax="proto3";
package {{clean .ModuleName}};

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
