package main

const (
	deliveryGRPCTemplate = `// {{.Header}}

package grpchandler

import (
	"context"
	"time"

	proto "{{.ProtoSource}}/{{.ModuleName}}"
	"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/usecase"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"

	"google.golang.org/grpc"

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
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
	mwGroup.AddProto(proto.File_{{cleanPathModule .ModuleName}}_{{cleanPathModule .ModuleName}}_proto, h.GetAll{{clean (upper .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{cleanPathModule .ModuleName}}_{{cleanPathModule .ModuleName}}_proto, h.GetDetail{{clean (upper .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{cleanPathModule .ModuleName}}_{{cleanPathModule .ModuleName}}_proto, h.Save{{clean (upper .ModuleName)}}, h.mw.GRPCBearerAuth)
}

// GetAll{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) GetAll{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.GetAll{{clean (upper .ModuleName)}}Request) (*proto.GetAll{{clean (upper .ModuleName)}}Response, error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	filter := candishared.Filter{
		Limit: int(req.Limit), Page: int(req.Page), Search: req.Search, OrderBy: req.OrderBy, Sort: req.Sort, ShowAll: req.ShowAll,
	}

	data, meta, err := h.uc.GetAll{{clean (upper .ModuleName)}}(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := &proto.GetAll{{clean (upper .ModuleName)}}Response{
		Meta: &proto.Meta{
			Page: int64(meta.Page), Limit: int64(meta.Limit), TotalRecords: int64(meta.TotalRecords), TotalPages: int64(meta.TotalPages),
		},
	}

	for _, d := range data {
		resp.Data = append(resp.Data, &proto.{{clean (upper .ModuleName)}}Model{
			ID: d.ID, CreatedAt: d.CreatedAt.Format(time.RFC3339), ModifiedAt: d.ModifiedAt.Format(time.RFC3339),
		})
	}

	return resp, nil
}

// GetDetail{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) GetDetail{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.GetDetail{{clean (upper .ModuleName)}}Request) (*proto.{{clean (upper .ModuleName)}}Model, error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:GetDetail{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	data, err := h.uc.GetDetail{{clean (upper .ModuleName)}}(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &proto.{{clean (upper .ModuleName)}}Model{
		ID: data.ID, CreatedAt: data.CreatedAt.Format(time.RFC3339), ModifiedAt: data.ModifiedAt.Format(time.RFC3339),
	}, nil
}

// Save{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) Save{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.{{clean (upper .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:Save{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	mErr := candihelper.NewMultiError()

	var payload shareddomain.{{clean (upper .ModuleName)}}
	payload.ID = req.ID

	payload.CreatedAt, err = time.Parse(time.RFC3339, req.CreatedAt)
	mErr.Append("createdAt", err)
	payload.ModifiedAt, err = time.Parse(time.RFC3339, req.ModifiedAt)
	mErr.Append("modifiedAt", err)

	if mErr.HasError() {
		return nil, mErr
	}

	if err := h.uc.Save{{clean (upper .ModuleName)}}(ctx, &payload); err != nil {
		return nil, err
	}

	return &proto.Response{
		Message: "Success",
	}, nil
}

// Delete{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) Delete{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.{{clean (upper .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace := tracer.StartTrace(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()
	ctx = trace.Context()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	if err := h.uc.Delete{{clean (upper .ModuleName)}}(ctx, req.ID); err != nil {
		return nil, err
	}

	return &proto.Response{
		Message: "Success",
	}, nil
}
`

	defaultGRPCProto = `syntax="proto3";
package {{clean .ModuleName}};
option go_package = "{{.PackagePrefix}}/api/proto/{{.ModuleName}}";

service {{clean (upper .ModuleName)}}Handler {
	rpc GetAll{{clean (upper .ModuleName)}}(GetAll{{clean (upper .ModuleName)}}Request) returns (GetAll{{clean (upper .ModuleName)}}Response);
	rpc GetDetail{{clean (upper .ModuleName)}}(GetDetail{{clean (upper .ModuleName)}}Request) returns ({{clean (upper .ModuleName)}}Model);
	rpc Save{{clean (upper .ModuleName)}}({{clean (upper .ModuleName)}}Model) returns (Response);
	rpc Delete{{clean (upper .ModuleName)}}({{clean (upper .ModuleName)}}Model) returns (Response);
}

message Meta {
	int64 Limit=1;
	int64 Page=2;
	int64 TotalRecords=3;
	int64 TotalPages=4;
}

message GetAll{{clean (upper .ModuleName)}}Request {
	int64 Limit=1;
	int64 Page=2;
	string Search=3;
	string OrderBy=4;
	string Sort=5;
	bool ShowAll=6;
}

message GetAll{{clean (upper .ModuleName)}}Response {
	Meta Meta=1;
	repeated {{clean (upper .ModuleName)}}Model Data=2;
}

message GetDetail{{clean (upper .ModuleName)}}Request {
	string ID=1;
}

message {{clean (upper .ModuleName)}}Model {
	string ID=1;
	string CreatedAt=2;
	string ModifiedAt=3;
}

message Response {
	string Message=1;
}
`
)
