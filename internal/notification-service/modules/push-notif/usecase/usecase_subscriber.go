package usecase

import (
	"context"
	"strconv"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
)

type helloSaidSubscriber struct {
	stop   <-chan struct{}
	events chan<- *domain.HelloSaidEvent
}

func (uc *pushNotifUsecaseImpl) runSubscriberListener() {
	subscribers := map[string]*helloSaidSubscriber{}
	unsubscribe := make(chan string)

	for {
		select {
		case id := <-unsubscribe:
			delete(subscribers, id)
		case s := <-uc.helloSaidSubscriber:
			subscribers[strconv.Itoa(int(time.Now().Unix()))] = s
		case e := <-uc.helloSaidEvents:
			for id, subs := range subscribers {
				go func(id string, subs *helloSaidSubscriber) {
					select {
					case <-subs.stop:
						unsubscribe <- id
						return
					default:
					}

					select {
					case <-subs.stop:
						unsubscribe <- id
					case subs.events <- e:
					case <-time.After(time.Second):
					}
				}(id, subs)
			}
		}
	}
}

func (uc *pushNotifUsecaseImpl) AddSubscriber(ctx context.Context) <-chan *domain.HelloSaidEvent {
	c := make(chan *domain.HelloSaidEvent)

	uc.helloSaidSubscriber <- &helloSaidSubscriber{
		events: c, stop: ctx.Done(),
	}

	return c
}
