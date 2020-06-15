package grpchandler

import (
	"context"
	"time"

	proto "agungdwiprasetyo.com/backend-microservices/api/line-chatbot/proto/event"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"

	"google.golang.org/grpc"
)

// GRPCHandler handler
type GRPCHandler struct {
	uc usecase.EventUsecase
}

// NewGRPCHandler create new rest handler
func NewGRPCHandler(uc usecase.EventUsecase) *GRPCHandler {
	return &GRPCHandler{uc}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server) {
	proto.RegisterEventHandlerServer(server, h)
}

// FindAll rpc
func (h *GRPCHandler) FindAll(ctx context.Context, payload *proto.Filter) (*proto.ResponseFindAll, error) {
	var filter shared.Filter
	filter.Limit = payload.Limit
	filter.OrderBy = payload.OrderBy
	filter.Page = payload.Page
	filter.Sort = payload.Sort

	data, meta, err := h.uc.FindAll(ctx, &filter)
	if err != nil {
		return nil, err
	}

	var response proto.ResponseFindAll
	response.Meta = &proto.Meta{
		Limit:        int32(meta.Limit),
		Page:         int32(meta.Page),
		TotalPages:   int64(meta.TotalPages),
		TotalRecords: int64(meta.TotalRecords),
	}
	for _, e := range data {
		response.Events = append(response.Events, &proto.Event{
			ID:         e.ID.Hex(),
			ReplyToken: e.ReplyToken,
			Type:       e.Type,
			Timestamp:  e.Timestamp.Format(time.RFC3339),
			SourceID:   e.SourceID,
			SourceType: e.SourceType,
			Message: &proto.Event_Message{
				ID:       e.Message.ID,
				Type:     e.Message.Type,
				Text:     e.Message.Text,
				Response: e.Message.Response,
			},
		})
	}

	return &response, nil
}
