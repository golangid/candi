package graphqlhandler

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

type subscriptionResolver struct {
	uc usecase.PushNotifUsecase
	mw interfaces.Middleware
}

func (s *subscriptionResolver) HelloSaid(ctx context.Context) <-chan *domain.HelloSaidEvent {
	return s.uc.AddSubscriber(ctx)
}
