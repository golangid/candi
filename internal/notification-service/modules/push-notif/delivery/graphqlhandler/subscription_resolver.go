package graphqlhandler

import (
	"context"
	"fmt"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

type subscriptionResolver struct {
	uc usecase.PushNotifUsecase
	mw interfaces.Middleware
}

func (s *subscriptionResolver) ListenTopic(ctx context.Context, input subscribeInputResolver) (<-chan *domain.Event, error) {

	tokenClaim, err := s.mw.Bearer(ctx, input.Token)
	if err != nil {
		logger.LogE(err.Error())
		return nil, err
	}

	clientID := tokenClaim.Audience
	fmt.Println(clientID)
	return s.uc.AddSubscriber(ctx, input.Token, input.Topic), nil
}
