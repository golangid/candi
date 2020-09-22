package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
)

var mutex sync.Mutex

func channelKey(topic, subscriberID string) string {
	return fmt.Sprintf("%s~%s", topic, subscriberID)
}

func (uc *pushNotifUsecaseImpl) AddSubscriber(ctx context.Context, clientID, topic string) <-chan *domain.Event {
	trace := tracer.StartTrace(ctx, "Usecase:AddSubscriber")
	defer trace.Finish()
	ctx = trace.Context()

	event := make(chan *domain.Event)

	go func() {
		newSubscriber := domain.Subscriber{ID: clientID, Topic: topic, IsActive: true}
		uc.registerNewSubscriberInTopic(ctx, &newSubscriber, event)

		select {
		case <-ctx.Done():
			close(event)
			uc.removeSubscriber(context.Background(), &newSubscriber)
			return
		}
	}()

	return event
}

func (uc *pushNotifUsecaseImpl) PublishMessageToTopic(ctx context.Context, event *domain.Event) (*domain.Event, error) {
	trace := tracer.StartTrace(ctx, "Usecase:PublishMessageToTopic")
	defer trace.Finish()
	ctx = trace.Context()
	tags := trace.Tags()

	tags["event"] = event

	repoRes := <-uc.repo.Subscriber.FindTopic(ctx, domain.Topic{Name: event.Topic})
	if repoRes.Error != nil {
		logger.LogE(repoRes.Error.Error())
		return nil, errors.New("Topic not found")
	}

	topic := repoRes.Data.(domain.Topic)

	tags["total_subscribers"] = len(topic.Subscribers)

	go func() {
		defer func() { recover() }()

		// broadcast event to subscriber topic
		for _, subs := range topic.Subscribers {
			subs.Topic = topic.Name
			subscriber, ok := uc.eventChannelSubscribers[channelKey(subs.Topic, subs.ID)]
			if !ok {
				logger.LogI("subscriber inactive")
				continue
			}

			subscriber <- event
		}
	}()

	// save log message

	return event, nil
}

func (uc *pushNotifUsecaseImpl) registerNewSubscriberInTopic(ctx context.Context, subscriber *domain.Subscriber, event chan<- *domain.Event) {
	topic := domain.Topic{Name: subscriber.Topic}
	repoRes := <-uc.repo.Subscriber.FindTopic(ctx, topic)
	if repoRes.Error != nil {
		logger.LogE(repoRes.Error.Error())
	} else {
		topic = repoRes.Data.(domain.Topic)
	}

	subscriber.ModifiedAt = time.Now()
	repoRes = <-uc.repo.Subscriber.FindSubscriber(ctx, subscriber.Topic, subscriber)
	if repoRes.Error != nil {
		subscriber.CreatedAt = time.Now()
		topic.Subscribers = append(topic.Subscribers, subscriber)
	} else {
		for i, subs := range topic.Subscribers {
			if subs.ID == subscriber.ID {
				topic.Subscribers[i].ModifiedAt = subscriber.ModifiedAt
				topic.Subscribers[i].IsActive = subscriber.IsActive
				break
			}
		}
	}

	if err := <-uc.repo.Subscriber.Save(ctx, &topic); err != nil {
		logger.LogE(err.Error())
	}

	mutex.Lock()
	defer mutex.Unlock()
	uc.eventChannelSubscribers[channelKey(subscriber.Topic, subscriber.ID)] = event
}

func (uc *pushNotifUsecaseImpl) removeSubscriber(ctx context.Context, subscriber *domain.Subscriber) {

	logger.LogIf("unsubscribe topic: %s, userID: %s", subscriber.Topic, subscriber.ID)

	subscriber.ModifiedAt = time.Now()
	uc.repo.Subscriber.RemoveSubscriber(ctx, subscriber)

	mutex.Lock()
	defer mutex.Unlock()
	delete(uc.eventChannelSubscribers, channelKey(subscriber.Topic, subscriber.ID))
}
