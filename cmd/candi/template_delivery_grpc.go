package main

const (
	deliveryGRPCTemplate = `// {{.Header}}

package grpchandler

import (
	"context"

	proto "{{.ProtoSource}}/{{.ModuleName}}"
	"{{$.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/domain"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
		StartDate: req.StartDate, EndDate: req.EndDate,
	}
	if err := h.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/get_all", filter); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	result, err := h.uc.{{upper (camel .ModuleName)}}().GetAll{{upper (camel .ModuleName)}}(ctx, &filter)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	resp := &proto.GetAll{{upper (camel .ModuleName)}}Response{
		Meta: &proto.Meta{
			Page: int64(result.Meta.Page), Limit: int64(result.Meta.Limit), TotalRecords: int64(result.Meta.TotalRecords), TotalPages: int64(result.Meta.TotalPages),
		},
	}

	for _, d := range result.Data {
		data := &proto.{{upper (camel .ModuleName)}}Model{
			Id: {{if and .MongoDeps (not .SQLDeps)}}d.ID{{else}}int64(d.ID){{end}}, Field: d.Field, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
		}
		resp.Data = append(resp.Data, data)
	}

	return resp, nil
}

// GetDetail{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) GetDetail{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.GetDetail{{upper (camel .ModuleName)}}Request) (*proto.{{upper (camel .ModuleName)}}Model, error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:GetDetail{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	data, err := h.uc.{{upper (camel .ModuleName)}}().GetDetail{{upper (camel .ModuleName)}}(ctx, {{if and .MongoDeps (not .SQLDeps)}}req.Id{{else}}int(req.Id){{end}})
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	resp := &proto.{{upper (camel .ModuleName)}}Model{
		Id: {{if and .MongoDeps (not .SQLDeps)}}data.ID{{else}}int64(data.ID){{end}}, Field: data.Field, CreatedAt: data.CreatedAt, UpdatedAt: data.UpdatedAt,
	}
	return resp, nil
}

// Create{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) Create{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.Request{{upper (camel .ModuleName)}}Model) (resp *proto.{{upper (camel .ModuleName)}}Model, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:Create{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	var payload domain.Request{{upper (camel .ModuleName)}}
	payload.Field = req.Field
	if err := h.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", payload); err != nil {
		return nil,  status.Errorf(codes.InvalidArgument, err.Error())
	}
	data, err := h.uc.{{upper (camel .ModuleName)}}().Create{{upper (camel .ModuleName)}}(ctx, &payload)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	resp = &proto.{{upper (camel .ModuleName)}}Model{
		Id: {{if and .MongoDeps (not .SQLDeps)}}data.ID{{else}}int64(data.ID){{end}}, Field: data.Field, CreatedAt: data.CreatedAt, UpdatedAt: data.UpdatedAt,
	}
	return resp, nil
}

// Update{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) Update{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.Request{{upper (camel .ModuleName)}}Model) (resp *proto.BaseResponse, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:Update{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	var payload domain.Request{{upper (camel .ModuleName)}}
	payload.ID = {{if and .MongoDeps (not .SQLDeps)}}req.Id{{else}}int(req.Id){{end}}
	payload.Field = req.Field
	if err := h.validator.ValidateDocument("{{cleanPathModule .ModuleName}}/save", payload); err != nil {
		return nil,  status.Errorf(codes.InvalidArgument, err.Error())
	}
	if err := h.uc.{{upper (camel .ModuleName)}}().Update{{upper (camel .ModuleName)}}(ctx, &payload); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	return &proto.BaseResponse{
		Message: "Success",
	}, nil
}

// Delete{{upper (camel .ModuleName)}} rpc method
func (h *GRPCHandler) Delete{{upper (camel .ModuleName)}}(ctx context.Context, req *proto.Request{{upper (camel .ModuleName)}}Model) (resp *proto.BaseResponse, err error) {
	trace, ctx := tracer.StartTraceWithContext(ctx, "{{upper (camel .ModuleName)}}DeliveryGRPC:Delete{{upper (camel .ModuleName)}}")
	defer trace.Finish()

	// tokenClaim := candishared.ParseTokenClaimFromContext(ctx) // must using GRPCBearerAuth in middleware for this handler

	if err := h.uc.{{upper (camel .ModuleName)}}().Delete{{upper (camel .ModuleName)}}(ctx, {{if and .MongoDeps (not .SQLDeps)}}req.Id{{else}}int(req.Id){{end}}); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	return &proto.BaseResponse{
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
	rpc Create{{upper (camel .ModuleName)}}(Request{{upper (camel .ModuleName)}}Model) returns ({{upper (camel .ModuleName)}}Model);
	rpc Update{{upper (camel .ModuleName)}}(Request{{upper (camel .ModuleName)}}Model) returns (BaseResponse);
	rpc Delete{{upper (camel .ModuleName)}}(Request{{upper (camel .ModuleName)}}Model) returns (BaseResponse);
}

message Meta {
	int64 limit=1;
	int64 page=2;
	int64 totalRecords=3;
	int64 totalPages=4;
}

message GetAll{{upper (camel .ModuleName)}}Request {
	int64 limit=1;
	int64 page=2;
	string search=3;
	string orderBy=4;
	string sort=5;
	bool showAll=6;
	string startDate=7;
	string endDate=8;
}

message GetAll{{upper (camel .ModuleName)}}Response {
	Meta meta=1;
	repeated {{upper (camel .ModuleName)}}Model data=2;
}

message GetDetail{{upper (camel .ModuleName)}}Request {
	{{if and .MongoDeps (not .SQLDeps)}}string{{else}}int64{{end}} id=1;
}

message Request{{upper (camel .ModuleName)}}Model {
	{{if and .MongoDeps (not .SQLDeps)}}string{{else}}int64{{end}} id=1;
	string field=2;
}

message {{upper (camel .ModuleName)}}Model {
	{{if and .MongoDeps (not .SQLDeps)}}string{{else}}int64{{end}} id=1;
	string field=2;
	string createdAt=3;
	string updatedAt=4;
}

message BaseResponse {
	string message=1;
}
`
)
