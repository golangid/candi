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

	"google.golang.org/grpc"` + `
	
	{{if and .MongoDeps (not .SQLDeps)}}"go.mongodb.org/mongo-driver/bson/primitive"{{end}}` + `

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
	mwGroup.AddProto(proto.File_{{cleanPathModule .ModuleName}}_{{cleanPathModule .ModuleName}}_proto, h.Create{{clean (upper .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{cleanPathModule .ModuleName}}_{{cleanPathModule .ModuleName}}_proto, h.Update{{clean (upper .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{cleanPathModule .ModuleName}}_{{cleanPathModule .ModuleName}}_proto, h.Delete{{clean (upper .ModuleName)}}, h.mw.GRPCBearerAuth)
}

// GetAll{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) GetAll{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.GetAll{{clean (upper .ModuleName)}}Request) (*proto.GetAll{{clean (upper .ModuleName)}}Response, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:GetAll{{clean (upper .ModuleName)}}")
	defer trace.Finish()

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
		data := &proto.{{clean (upper .ModuleName)}}Model{
			CreatedAt: d.CreatedAt.Format(time.RFC3339), ModifiedAt: d.ModifiedAt.Format(time.RFC3339),
		}
		data.ID = d.ID{{if and .MongoDeps (not .SQLDeps)}}.Hex(){{end}}
		resp.Data = append(resp.Data, data)
	}

	return resp, nil
}

// GetDetail{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) GetDetail{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.GetDetail{{clean (upper .ModuleName)}}Request) (*proto.{{clean (upper .ModuleName)}}Model, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:GetDetail{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	data, err := h.uc.GetDetail{{clean (upper .ModuleName)}}(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	resp := &proto.{{clean (upper .ModuleName)}}Model{
		CreatedAt: data.CreatedAt.Format(time.RFC3339), ModifiedAt: data.ModifiedAt.Format(time.RFC3339),
	}
	resp.ID = data.ID{{if and .MongoDeps (not .SQLDeps)}}.Hex(){{end}}
	return resp, nil
}

// Create{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) Create{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.{{clean (upper .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:Create{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	mErr := candihelper.NewMultiError()

	var payload shareddomain.{{clean (upper .ModuleName)}}
	{{if and .MongoDeps (not .SQLDeps)}}payload.ID, err = primitive.ObjectIDFromHex(req.ID)
	mErr.Append("id", err){{else}}payload.ID = req.ID{{end}}

	payload.CreatedAt, err = time.Parse(time.RFC3339, req.CreatedAt)
	mErr.Append("createdAt", err)
	payload.ModifiedAt, err = time.Parse(time.RFC3339, req.ModifiedAt)
	mErr.Append("modifiedAt", err)
	if mErr.HasError() {
		return nil, mErr
	}

	if err := h.uc.Create{{clean (upper .ModuleName)}}(ctx, &payload); err != nil {
		return nil, err
	}

	return &proto.Response{
		Message: "Success",
	}, nil
}

// Update{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) Update{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.{{clean (upper .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:Update{{clean (upper .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	var payload shareddomain.{{clean (upper .ModuleName)}}
	{{if and .MongoDeps (not .SQLDeps)}}if payload.ID, err = primitive.ObjectIDFromHex(req.ID); err != nil {
		return nil, err
	}{{else}}payload.ID = req.ID{{end}}

	if err := h.uc.Update{{clean (upper .ModuleName)}}(ctx, req.ID, &payload); err != nil {
		return nil, err
	}

	return &proto.Response{
		Message: "Success",
	}, nil
}

// Delete{{clean (upper .ModuleName)}} rpc method
func (h *GRPCHandler) Delete{{clean (upper .ModuleName)}}(ctx context.Context, req *proto.{{clean (upper .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{clean (upper .ModuleName)}}DeliveryGRPC:Delete{{clean (upper .ModuleName)}}")
	defer trace.Finish()

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
	rpc Create{{clean (upper .ModuleName)}}({{clean (upper .ModuleName)}}Model) returns (Response);
	rpc Update{{clean (upper .ModuleName)}}({{clean (upper .ModuleName)}}Model) returns (Response);
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
	string Field=2;
	string CreatedAt=3;
	string ModifiedAt=4;
}

message Response {
	string Message=1;
}
`
)
