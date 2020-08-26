package graphqlhandler

import (
	"context"
	"errors"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

type mutationResolver struct {
	uc usecase.PushNotifUsecase
	mw interfaces.Middleware
}

// Push resolver
func (m *mutationResolver) Push(ctx context.Context, input pushInputResolver) (string, error) {
	trace := tracer.StartTrace(ctx, "Delivery-Push")
	defer trace.Finish()
	m.mw.GraphQLBasicAuth(ctx)

	ctx, tags := trace.Context(), trace.Tags()
	tags["input"] = input

	if err := m.uc.SendNotification(ctx, input.Payload); err != nil {
		return "", err
	}

	return "Ok", nil
}

// ScheduledNotification resolver
func (m *mutationResolver) ScheduledNotification(ctx context.Context, input scheduleNotifInputResolver) (string, error) {
	trace := tracer.StartTrace(ctx, "Delivery-ScheduledNotification")
	defer trace.Finish()
	m.mw.GraphQLBasicAuth(ctx)

	ctx = trace.Context()

	scheduledAt, err := time.Parse(time.RFC3339, input.Payload.ScheduledAt)
	if err != nil {
		return "Failed parse scheduled time format", err
	}

	if scheduledAt.Before(time.Now()) {
		return "", errors.New("Scheduled time must in future")
	}

	if err := m.uc.SendScheduledNotification(ctx, scheduledAt, input.Payload.Data); err != nil {
		return "Failed set scheduled push notification", err
	}

	return "Success set scheduled push notification, scheduled at: " + input.Payload.ScheduledAt, nil
}

// ScheduledBroadcastEvent resolver
func (m *mutationResolver) ScheduledBroadcastEvent(ctx context.Context, input struct {
	ScheduledAt string
	Event       inputTopicEvent
}) (string, error) {
	trace := tracer.StartTrace(ctx, "Delivery-ScheduledBroadcastEvent")
	defer trace.Finish()
	// m.mw.GraphQLBasicAuth(ctx)

	ctx = trace.Context()

	scheduledAt, err := time.Parse(time.RFC3339, input.ScheduledAt)
	if err != nil {
		return "Failed parse scheduled time format", err
	}

	if scheduledAt.Before(time.Now()) {
		return "", errors.New("Scheduled time must in future")
	}

	if err := m.uc.SendScheduledEvent(ctx, scheduledAt, &domain.Event{
		ID: input.Event.ID, Message: input.Event.Message, Topic: input.Event.Topic,
	}); err != nil {
		return "Failed set scheduled event", err
	}

	return "Success set scheduled event, scheduled at: " + input.ScheduledAt, nil
}

func (m *mutationResolver) PublishMessageToTopic(ctx context.Context, input *inputTopicEvent) *domain.Event {

	tokenClaim := m.mw.GraphQLBearerAuth(ctx)

	e := &domain.Event{Message: input.Message, Topic: input.Topic, ID: tokenClaim.User.Username}
	e, _ = m.uc.PublishMessageToTopic(ctx, e)
	return e
}
