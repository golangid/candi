package cronworker

// cron scheduler worker, create with 100% pure internal go library (using reflect select channel)

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

type cronWorker struct {
	service factory.ServiceFactory
	wg      sync.WaitGroup
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	refreshWorkerNotif, shutdown = make(chan struct{}), make(chan struct{})
	semaphore = make(chan struct{}, config.BaseEnv().MaxGoroutines)

	// add shutdown channel to first index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(shutdown),
	})
	// add refresh worker channel to second index
	workers = append(workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(refreshWorkerNotif),
	})

	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.Scheduler); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				funcName, interval := candihelper.ParseCronJobKey(handler.Pattern)

				var job Job
				job.HandlerName = funcName
				job.HandlerFunc = handler.HandlerFunc
				job.Interval = interval
				if err := AddJob(job); err != nil {
					panic(err)
				}

				logger.LogYellow(fmt.Sprintf(`[CRON-WORKER] job_name: %s~%s -> every: %s`, m.Name(), funcName, interval))
			}
		}
	}
	fmt.Printf("\x1b[34;1m⇨ Cron worker running with %d jobs\x1b[0m\n\n", len(activeJobs))

	return &cronWorker{
		service: service,
	}
}

func (c *cronWorker) Serve() {

	for {
		chosen, _, ok := reflect.Select(workers)
		if !ok {
			continue
		}

		// if shutdown channel captured, break loop (no more jobs will run)
		if chosen == 0 {
			break
		}

		// notify for refresh worker
		if chosen == 1 {
			continue
		}

		chosen = chosen - 2
		job := activeJobs[chosen]
		if job.nextDuration != nil {
			job.ticker.Stop()
			job.ticker = time.NewTicker(*job.nextDuration)
			workers[job.WorkerIndex].Chan = reflect.ValueOf(job.ticker.C)
			activeJobs[chosen].nextDuration = nil
		}

		c.wg.Add(1)
		semaphore <- struct{}{}
		go func(job *Job) {
			defer c.wg.Done()

			trace := tracer.StartTrace(context.Background(), "CronScheduler")
			defer trace.Finish()
			ctx := trace.Context()

			defer func() {
				if r := recover(); r != nil {
					trace.SetError(fmt.Errorf("%v", r))
				}
				<-semaphore
				logger.LogGreen("cron scheduler " + tracer.GetTraceURL(ctx))
			}()

			if config.BaseEnv().DebugMode {
				log.Printf("\x1b[35;3mCron Scheduler: executing task '%s' (interval: %s)\x1b[0m", job.HandlerName, job.Interval)
			}

			tags := trace.Tags()
			tags["job_name"] = job.HandlerName
			if err := job.HandlerFunc(ctx, []byte(job.Params)); err != nil {
				trace.SetError(err)
			}
		}(job)

	}
}

func (c *cronWorker) Shutdown(ctx context.Context) {
	log.Println("\x1b[33;1mStopping Cron Job Scheduler worker...\x1b[0m")
	defer func() { log.Println("\x1b[33;1mStopping Cron Job Scheduler:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m") }()

	if len(activeJobs) == 0 {
		return
	}

	shutdown <- struct{}{}
	runningJob := len(semaphore)
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mCron Job Scheduler:\x1b[0m waiting %d job until done...\n", runningJob)
	}

	c.wg.Wait()
}
