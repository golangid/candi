package main

const (
	deliveryGRPCTemplate = `package grpchandler

import (
	"context"

	proto "{{.PackageName}}/api/{{.ServiceName}}/proto/{{$.module}}"
	"{{.PackageName}}/pkg/codebase/interfaces"
	"google.golang.org/grpc"
)

// GRPCHandler rpc handler
type GRPCHandler struct {
	mw interfaces.Middleware
}

// NewGRPCHandler func
func NewGRPCHandler(mw interfaces.Middleware) *GRPCHandler {
	return &GRPCHandler{
		mw: mw,
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server) {
	proto.Register{{clean (upper $.module)}}HandlerServer(server, h)
}

// FindAll rpc
func (h *GRPCHandler) Hello(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	return &proto.Response{
		Message: req.Message + "; Hello, from service: {{$.ServiceName}}, module: {{$.module}}",
	}, nil
}

`

	defaultGRPCProto = `syntax="proto3";
package {{clean $.module}};

service {{clean (upper $.module)}}Handler {
	rpc Hello(Request) returns (Response);
}

message Request {
    string Message=1;
}

message Response {
	string Message=1;
}`
)
