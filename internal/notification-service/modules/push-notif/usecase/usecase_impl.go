package usecase

import (
	"context"
	"os"

	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/pushnotif"
)

type pushNotifUsecaseImpl struct {
	pushNotif pushnotif.PushNotif
}

// NewPushNotifUsecase constructor
func NewPushNotifUsecase() PushNotifUsecase {
	return &pushNotifUsecaseImpl{
		pushNotif: pushnotif.NewFirebaseREST(
			os.Getenv("FIREBASE_HOST"), os.Getenv("FIREBASE_KRAB_KEY"),
		),
	}
}

func (uc *pushNotifUsecaseImpl) SendNotification(ctx context.Context) (err error) {

	requestPayload := pushnotif.PushRequest{
		Notification: &pushnotif.Notification{
			Title:          "Testing",
			Body:           "Hello",
			Image:          "https://storage.googleapis.com/agungdp/static/logo/golang.png",
			Sound:          "default",
			MutableContent: true,
			ResourceID:     "resourceID",
			ResourceName:   "resourceName",
		},
		Data: map[string]interface{}{"type": "type"},
	}
	result := <-uc.pushNotif.Push(ctx, requestPayload)
	if result.Error != nil {
		return result.Error
	}

	logger.LogI("success send notification")
	return
}
