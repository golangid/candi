package usecase

import (
	"context"
	"encoding/json"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

type pushNotifUsecaseImpl struct {
	modName types.Module
	repo    *repository.Repository

	// for subscriber listener
	helloSaidEvents     chan *domain.HelloSaidEvent
	helloSaidSubscriber chan *helloSaidSubscriber
}

// NewPushNotifUsecase constructor
func NewPushNotifUsecase(modName types.Module, repo *repository.Repository) PushNotifUsecase {
	helloEvent := make(chan *domain.HelloSaidEvent)
	helloSubscriber := make(chan *helloSaidSubscriber)

	uc := &pushNotifUsecaseImpl{
		modName: modName,
		repo:    repo,

		helloSaidEvents:     helloEvent,
		helloSaidSubscriber: helloSubscriber,
	}

	go uc.runSubscriberListener()

	return uc
}

func (uc *pushNotifUsecaseImpl) SendNotification(ctx context.Context, request *domain.PushNotifRequestPayload) (err error) {
	trace := tracer.StartTrace(ctx, "Usecase-SendNotification")
	defer trace.Finish()
	ctx = trace.Context()

	requestPayload := domain.PushRequest{
		To: request.To,
		Notification: &domain.Notification{
			Title:          request.Title,
			Body:           request.Message,
			Image:          "https://storage.googleapis.com/agungdp/static/logo/golang.png",
			Sound:          "default",
			MutableContent: true,
			ResourceID:     "resourceID",
			ResourceName:   "resourceName",
		},
		Data: map[string]interface{}{"type": "type"},
	}
	result := <-uc.repo.PushNotif.Push(ctx, requestPayload)
	if result.Error != nil {
		return result.Error
	}

	logger.LogI("success send notification")
	return
}

func (uc *pushNotifUsecaseImpl) SendScheduledNotification(ctx context.Context, scheduledAt time.Time, request *domain.PushNotifRequestPayload) (err error) {
	trace := tracer.StartTrace(ctx, "Usecase-SendScheduledNotification")
	defer trace.Finish()
	ctx = trace.Context()

	redisTopicKey := helper.BuildRedisPubSubKeyTopic(string(uc.modName), "scheduled-push-notif")
	data, _ := json.Marshal(request)
	exp := scheduledAt.Sub(time.Now())
	return uc.repo.Schedule.SaveScheduledNotification(ctx, redisTopicKey, data, exp)
}

func (uc *pushNotifUsecaseImpl) SayHello(ctx context.Context, event *domain.HelloSaidEvent) *domain.HelloSaidEvent {
	go func() {
		select {
		case uc.helloSaidEvents <- event:
		case <-time.After(1 * time.Second):
		}
	}()
	return event
}
