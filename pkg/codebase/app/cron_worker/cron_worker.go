package cronworker

// cron scheduler worker, create with 100% pure internal go library (using reflect select channel)

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

type cronWorker struct {
	service factory.ServiceFactory
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	return &cronWorker{
		service: service,
	}
}

func (c *cronWorker) Serve() {
	var jobs []schedulerJob
	var schedulerChannels []reflect.SelectCase
	for _, m := range c.service.GetModules() {
		if h := m.WorkerHandler(constant.Scheduler); h != nil {
			for _, topic := range h.GetTopics() {
				funcName, interval := helper.ParseCronJobKey(topic)
				duration, err := time.ParseDuration(interval)
				if err != nil {
					panic(err)
				}

				job := schedulerJob{
					handlerName: funcName,
					ticker:      time.NewTicker(duration),
					handler:     h,
				}

				schedulerChannels = append(schedulerChannels, reflect.SelectCase{
					Dir: reflect.SelectRecv, Chan: reflect.ValueOf(job.ticker.C),
				})
				jobs = append(jobs, job)
			}
		}
	}

	if len(jobs) == 0 {
		log.Println("No scheduler handler found")
		return
	}

	fmt.Printf("\x1b[34;1m⇨ Cron worker running with %d jobs\x1b[0m\n\n", len(schedulerChannels))
	for {
		chosen, _, ok := reflect.Select(schedulerChannels)
		if !ok {
			continue
		}

		go func(selected int) {
			defer func() { recover() }()
			job := jobs[selected]
			job.handler.ProcessMessage(context.Background(), job.handlerName, []byte(job.params))
		}(chosen)
	}
}

func (c *cronWorker) Shutdown(ctx context.Context) {
	log.Println("Stopping cron job scheduler...")
	// TODO: handling graceful stop all channel from jobs ticker
}

type schedulerJob struct {
	ticker      *time.Ticker
	handlerName string
	handler     interfaces.WorkerHandler
	params      string
}
