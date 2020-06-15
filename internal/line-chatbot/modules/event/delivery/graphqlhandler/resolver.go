package graphqlhandler

import (
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

type EventResolver struct {
	e       domain.Event
	message EventMessage
}

func (r *EventResolver) ID() string {
	return r.e.ID.Hex()
}
func (r *EventResolver) ReplyToken() string {
	return r.e.ReplyToken
}
func (r *EventResolver) Type() string {
	return r.e.Type
}
func (r *EventResolver) Timestamp() string {
	return r.e.Timestamp.In(helper.AsiaJakartaLocalTime).Format(time.RFC3339)
}
func (r *EventResolver) SourceId() string {
	return r.e.SourceID
}
func (r *EventResolver) SourceType() string {
	return r.e.SourceType
}
func (r *EventResolver) Message() *EventMessage {
	return &r.message
}
func (r *EventResolver) Error() *string {
	return r.e.Error
}

type EventMessage struct {
	e domain.Event
}

func (r *EventMessage) ID() string {
	return r.e.Message.ID
}
func (r *EventMessage) Type() string {
	return r.e.Message.Type
}
func (r *EventMessage) Text() string {
	return r.e.Message.Text
}
func (r *EventMessage) Response() string {
	return r.e.Message.Response
}

type EventListResolver struct {
	m      *shared.Meta
	events []*EventResolver
}

func (r *EventListResolver) Meta() *shared.MetaResolver {
	return r.m.ToResolver()
}
func (r *EventListResolver) Data() []*EventResolver {
	return r.events
}
