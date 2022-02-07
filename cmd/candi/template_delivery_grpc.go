package main

const (
	deliveryGRPCTemplate = `// {{.Header}}

package grpchandler

import (
	"context"
	"time"

	proto "{{.ProtoSource}}/{{.ModuleName}}"
	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	shareddomain "{{$.PackagePrefix}}/pkg/shared/domain"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"google.golang.org/grpc"` + `
	
	{{if and .MongoDeps (not .SQLDeps)}}"go.mongodb.org/mongo-driver/bson/primitive"{{end}}` + `

	"{{.LibraryName}}/candihelper"
	"{{.LibraryName}}/candishared"
	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
	"{{.LibraryName}}/tracer"
)

// GRPCHandler rpc handler
type GRPCHandler struct {
	mw        interfaces.Middleware
	uc        usecase.Usecase
	validator interfaces.Validator
}

// NewGRPCHandler func
func NewGRPCHandler(uc usecase.Usecase, deps dependency.Dependency) *GRPCHandler {
	return &GRPCHandler{
		uc: uc, mw: deps.GetMiddleware(), validator: deps.GetValidator(),
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server, mwGroup *types.MiddlewareGroup) {
	proto.Register{{upper (camel .ModuleName)}}HandlerServer(server, h)

	// register middleware for method
	mwGroup.AddProto(proto.File_{{snake .ModuleName}}_{{snake .ModuleName}}_proto, h.GetAll{{upper (camel .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{snake .ModuleName}}_{{snake .ModuleName}}_proto, h.GetDetail{{upper (camel .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{snake .ModuleName}}_{{snake .ModuleName}}_proto, h.Create{{upper (camel .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{snake .ModuleName}}_{{snake .ModuleName}}_proto, h.Update{{upper (camel .ModuleName)}}, h.mw.GRPCBearerAuth)
	mwGroup.AddProto(proto.File_{{snake .ModuleName}}_{{snake .ModuleName}}_proto, h.Delete{{upper (camel .ModuleName)}}, h.mw.GRPCBearerAuth)
}

// GetAll{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) GetAll{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.GetAll{{upper (camel .ModuleName)}}Request) (*proto.GetAll{{upper (camel .ModuleName)}}Response, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:GetAll{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	filter := domain.Filter{{upper (camel .ModuleName)}}{
		Filter: candishared.Filter{
			Limit: int(req.Limit), Page: int(req.Page), Search: req.Search, OrderBy: req.OrderBy, Sort: req.Sort, ShowAll: req.ShowAll,
		},
	}

	data, meta, err := h.uc.{{upper (camel .ModuleName)}}().GetAll{{upper (camel .ModuleName)}}(ctx, &filter)
	if err != nil {
		return nil, err
	}

	resp := &proto.GetAll{{upper (camel .ModuleName)}}Response{
		Meta: &proto.Meta{
			Page: int64(meta.Page), Limit: int64(meta.Limit), TotalRecords: int64(meta.TotalRecords), TotalPages: int64(meta.TotalPages),
		},
	}

	for _, d := range data {
		data := &proto.{{upper (camel .ModuleName)}}Model{
			CreatedAt: d.CreatedAt.Format(time.RFC3339), UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
		}
		data.ID = d.ID{{if and .MongoDeps (not .SQLDeps)}}.Hex(){{end}}
		resp.Data = append(resp.Data, data)
	}

	return resp, nil
}

// GetDetail{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.GetDetail{{upper (camel .ModuleName)}}Request) (*proto.{{upper (camel .ModuleName)}}Model, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:GetDetail{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	data, err := h.uc.{{upper (camel .ModuleName)}}().GetDetail{{upper (camel .ModuleName)}}(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	resp := &proto.{{upper (camel .ModuleName)}}Model{
		CreatedAt: data.CreatedAt.Format(time.RFC3339), UpdatedAt: data.UpdatedAt.Format(time.RFC3339),
	}
	resp.ID = data.ID{{if and .MongoDeps (not .SQLDeps)}}.Hex(){{end}}
	return resp, nil
}

// Create{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) Create{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.{{upper (camel .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	mErr := candihelper.NewMultiError()

	var payload shareddomain.{{upper (camel .ModuleName)}}
	{{if and .MongoDeps (not .SQLDeps)}}payload.ID, err = primitive.ObjectIDFromHex(req.ID)
	mErr.Append("id", err){{else}}payload.ID = req.ID{{end}}

	payload.CreatedAt, err = time.Parse(time.RFC3339, req.CreatedAt)
	mErr.Append("createdAt", err)
	payload.UpdatedAt, err = time.Parse(time.RFC3339, req.UpdatedAt)
	mErr.Append("modifiedAt", err)
	if mErr.HasError() {
		return nil, mErr
	}

	if err := h.uc.{{upper (camel .ModuleName)}}().Create{{upper (camel .ModuleName)}}(ctx, &payload); err != nil {
		return nil, err
	}

	return &proto.Response{
		Message: "Success",
	}, nil
}

// Update{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) Update{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.{{upper (camel .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	var payload shareddomain.{{upper (camel .ModuleName)}}
	{{if and .MongoDeps (not .SQLDeps)}}if payload.ID, err = primitive.ObjectIDFromHex(req.ID); err != nil {
		return nil, err
	}{{else}}payload.ID = req.ID{{end}}

	if err := h.uc.{{upper (camel .ModuleName)}}().Update{{upper (camel .ModuleName)}}(ctx, req.ID, &payload); err != nil {
		return nil, err
	}

	return &proto.Response{
		Message: "Success",
	}, nil
}

// Delete{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) Delete{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.{{upper (camel .ModuleName)}}Model) (resp *proto.Response, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	if err := h.uc.{{upper (camel .ModuleName)}}().Delete{{upper (camel .ModuleName)}}(ctx, req.ID); err != nil {
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

service {{upper (camel .ModuleName)}}Handler {
	rpc GetAll{{upper (camel .ModuleName)}}(GetAll{{upper (camel .ModuleName)}}Request) returns (GetAll{{upper (camel .ModuleName)}}Response);
	rpc GetDetail{{upper (camel .ModuleName)}}(GetDetail{{upper (camel .ModuleName)}}Request) returns ({{upper (camel .ModuleName)}}Model);
	rpc Create{{upper (camel .ModuleName)}}({{upper (camel .ModuleName)}}Model) returns (Response);
	rpc Update{{upper (camel .ModuleName)}}({{upper (camel .ModuleName)}}Model) returns (Response);
	rpc Delete{{upper (camel .ModuleName)}}({{upper (camel .ModuleName)}}Model) returns (Response);
}

message Meta {
	int64 Limit=1;
	int64 Page=2;
	int64 TotalRecords=3;
	int64 TotalPages=4;
}

message GetAll{{upper (camel .ModuleName)}}Request {
	int64 Limit=1;
	int64 Page=2;
	string Search=3;
	string OrderBy=4;
	string Sort=5;
	bool ShowAll=6;
}

message GetAll{{upper (camel .ModuleName)}}Response {
	Meta Meta=1;
	repeated {{upper (camel .ModuleName)}}Model Data=2;
}

message GetDetail{{upper (camel .ModuleName)}}Request {
	string ID=1;
}

message {{upper (camel .ModuleName)}}Model {
	string ID=1;
	string Field=2;
	string CreatedAt=3;
	string UpdatedAt=4;
}

message Response {
	string Message=1;
}
`
)
