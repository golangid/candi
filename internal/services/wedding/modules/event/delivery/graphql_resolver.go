package delivery

import "github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/domain"

type EventResolver struct {
	e *domain.Event
}

func (r *EventResolver) ID() string {
	return r.e.ID
}
func (r *EventResolver) Code() string {
	return r.e.Code
}
func (r *EventResolver) Date() string {
	return r.e.Date
}
func (r *EventResolver) CountDown() int32 {
	return int32(r.e.CountDown)
}
func (r *EventResolver) Ceremony() string {
	return r.e.Ceremony
}
func (r *EventResolver) Reception() string {
	return r.e.Reception
}
func (r *EventResolver) Address() string {
	return r.e.Address
}
