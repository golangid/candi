package graphqlhandler

import "agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"

type pushInputResolver struct {
	Payload *domain.PushNotifRequestPayload
}

type scheduleNotifInputResolver struct {
	Payload struct {
		ScheduledAt string
		Data        *domain.PushNotifRequestPayload
	}
}
