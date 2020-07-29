package usecase

import (
	"context"
	"fmt"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

func (uc *pushNotifUsecaseImpl) runSubscriberListener() {

	subsChannel := make(map[string]*domain.Subscriber)

	for {
		select {
		case leave := <-uc.closer:
			logger.LogIf("unsubscribe topic: %s, userID: %s", leave.Topic, leave.ID)
			uc.repo.Subscriber.RemoveSubscriber(context.Background(), leave)
			delete(subsChannel, fmt.Sprintf("%s~%s", leave.Topic, leave.ID))

		case subs := <-uc.subscribers:

			newSubscriber := domain.Subscriber{ID: subs.ID, Topic: subs.Topic, IsActive: true}
			go uc.registerNewSubscriberInTopic(context.Background(), subs.Topic, &newSubscriber)
			key := fmt.Sprintf("%s~%s", newSubscriber.Topic, newSubscriber.ID)
			fmt.Println(key)
			subsChannel[key] = subs

		case e := <-uc.events:

			repoRes := <-uc.repo.Subscriber.FindTopic(context.Background(), domain.Topic{Name: e.ToTopic})
			if repoRes.Error != nil {
				logger.LogE(repoRes.Error.Error())
				continue
			}

			topic := repoRes.Data.(domain.Topic)
			for _, subs := range topic.Subscribers {
				subs.Topic = topic.Name
				go func(subs *domain.Subscriber) {
					subscriber, ok := subsChannel[fmt.Sprintf("%s~%s", subs.Topic, subs.ID)]
					if !ok {
						return
					}

					subscriber.Events <- e
				}(subs)
			}
		}
	}
}

func (uc *pushNotifUsecaseImpl) AddSubscriber(ctx context.Context, clientID, topic string) <-chan *domain.Event {
	event := make(chan *domain.Event)

	newSubs := &domain.Subscriber{
		ID:     clientID,
		Topic:  topic,
		Events: event,
	}

	uc.subscribers <- newSubs

	go func() {
		select {
		case <-ctx.Done():
			uc.closer <- newSubs
		}
	}()

	return event
}

func (uc *pushNotifUsecaseImpl) PublishMessageToTopic(ctx context.Context, event *domain.Event) *domain.Event {
	uc.events <- event

	// save message to database

	return event
}

func (uc *pushNotifUsecaseImpl) registerNewSubscriberInTopic(ctx context.Context, topicName string, subscriber *domain.Subscriber) {
	topic := domain.Topic{Name: topicName}
	repoRes := <-uc.repo.Subscriber.FindTopic(ctx, topic)
	if repoRes.Error != nil {
		logger.LogE(repoRes.Error.Error())
	} else {
		topic = repoRes.Data.(domain.Topic)
	}

	repoRes = <-uc.repo.Subscriber.FindSubscriber(ctx, topicName, subscriber)
	if repoRes.Error != nil {
		subscriber.Topic = topicName
		topic.Subscribers = append(topic.Subscribers, subscriber)
	}

	if err := <-uc.repo.Subscriber.Save(ctx, &topic); err != nil {
		logger.LogE(err.Error())
	}
}
