package cronworker

// cron scheduler worker, create with 100% pure internal go library (using reflect select channel)

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

type cronWorker struct {
	service   factory.ServiceFactory
	isHaveJob bool
	shutdown  chan struct{}
	wg        sync.WaitGroup
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	return &cronWorker{
		service:  service,
		shutdown: make(chan struct{}),
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
		log.Println("cronjob: no scheduler handler found")
		return
	}

	c.isHaveJob = true

	// add shutdown channel to last index
	schedulerChannels = append(schedulerChannels, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.shutdown),
	})

	fmt.Printf("\x1b[34;1m⇨ Cron worker running with %d jobs\x1b[0m\n\n", len(jobs))
	for {
		chosen, _, ok := reflect.Select(schedulerChannels)
		if !ok {
			continue
		}

		// if shutdown channel captured, break loop (no more jobs will run)
		if chosen == len(schedulerChannels)-1 {
			break
		}

		c.wg.Add(1)
		go func(selected int) {
			defer c.wg.Done()
			defer func() { recover() }()

			job := jobs[selected]
			job.handler.ProcessMessage(context.Background(), job.handlerName, []byte(job.params))
		}(chosen)
	}
}

func (c *cronWorker) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping cron job scheduler worker...")
	defer deferFunc()

	if !c.isHaveJob {
		return
	}

	c.shutdown <- struct{}{}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		fmt.Print("cronjob: force shutdown ")
	case <-done:
		fmt.Print("cronjob: success waiting all job until done ")
	}
}

type schedulerJob struct {
	ticker      *time.Ticker
	handlerName string
	handler     interfaces.WorkerHandler
	params      string
}
