package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Subscriber data storage
type Subscriber interface {
	Save(ctx context.Context, data *domain.Topic) <-chan error
	FindTopic(ctx context.Context, where domain.Topic) <-chan shared.Result
	RemoveSubscriber(ctx context.Context, subscriber *domain.Subscriber) <-chan error
	FindSubscriber(ctx context.Context, topicName string, subscriber *domain.Subscriber) <-chan shared.Result
}
