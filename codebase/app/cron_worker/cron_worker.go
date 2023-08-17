package cronworker

// cron scheduler worker, create with 100% pure internal go library (using reflect select channel)

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	cronexpr "github.com/golangid/candi/candiutils/cronparser"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

type cronWorker struct {
	ctx           context.Context
	ctxCancelFunc func()

	opt                          option
	service                      factory.ServiceFactory
	workers                      []reflect.SelectCase
	refreshWorkerNotif, shutdown chan struct{}
	semaphore                    []chan struct{}
	wg                           sync.WaitGroup
	activeJobs                   []*Job
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	c := &cronWorker{
		service: service,
		opt:     getDefaultOption(service),

		refreshWorkerNotif: make(chan struct{}),
		shutdown:           make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&c.opt)
	}

	// add shutdown channel to first index
	c.workers = append(c.workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.shutdown),
	})
	// add refresh worker channel to second index
	c.workers = append(c.workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.refreshWorkerNotif),
	})

	c.opt.locker.Reset(fmt.Sprintf(lockPattern, c.service.Name(), "*"))
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.Scheduler); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				funcName, args, interval := ParseCronJobKey(handler.Pattern)

				var job Job
				job.HandlerName = funcName
				job.Handler = handler
				job.Interval = interval
				job.Params = args
				if err := c.addJob(&job); err != nil {
					panic(fmt.Errorf(`Cron Worker: "%s" %v`, interval, err))
				}

				c.semaphore = append(c.semaphore, make(chan struct{}, c.opt.maxGoroutines))
				logger.LogYellow(fmt.Sprintf(`[CRON-WORKER] (job name): %s (every): %-8s  --> (module): "%s"`, `"`+funcName+`"`, interval, m.Name()))
			}
		}
	}
	fmt.Printf("\x1b[34;1m⇨ Cron worker running with %d jobs\x1b[0m\n\n", len(c.activeJobs))

	c.ctx, c.ctxCancelFunc = context.WithCancel(context.Background())
	return c
}

func (c *cronWorker) Serve() {

	for _, job := range c.activeJobs {
		c.workers[job.WorkerIndex].Chan = reflect.ValueOf(job.ticker.C)
	}

	// run worker
	for {
		chosen, _, ok := reflect.Select(c.workers)
		if !ok {
			continue
		}

		// if shutdown channel captured, break loop (no more jobs will run)
		if chosen == 0 {
			return
		}

		// notify for refresh worker
		if chosen == 1 {
			continue
		}

		chosen = chosen - 2
		job := c.activeJobs[chosen]
		c.registerNextInterval(job)

		if len(c.semaphore[job.WorkerIndex-2]) >= c.opt.maxGoroutines {
			continue
		}

		c.semaphore[job.WorkerIndex-2] <- struct{}{}
		c.wg.Add(1)
		go func(j *Job) {
			defer func() {
				c.wg.Done()
				<-c.semaphore[j.WorkerIndex-2]
			}()
			if c.ctx.Err() != nil {
				logger.LogRed("cron_scheduler > ctx root err: " + c.ctx.Err().Error())
				return
			}

			c.processJob(j)
		}(job)
	}

}

func (c *cronWorker) Shutdown(ctx context.Context) {
	defer func() {
		log.Println("\x1b[33;1mStopping Cron Job Scheduler:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")
	}()

	if len(c.activeJobs) == 0 {
		return
	}

	c.stopAllJob()
	c.shutdown <- struct{}{}
	runningJob := 0
	for _, sem := range c.semaphore {
		runningJob += len(sem)
	}
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mCron Job Scheduler:\x1b[0m waiting %d job until done...\n", runningJob)
	}

	c.wg.Wait()
	c.ctxCancelFunc()
	c.opt.locker.Reset(fmt.Sprintf(lockPattern, c.service.Name(), "*"))
}

func (c *cronWorker) Name() string {
	return string(types.Scheduler)
}

func (c *cronWorker) processJob(job *Job) {
	ctx := c.ctx
	if job.Handler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	// lock for multiple worker (if running on multiple pods/instance)
	if c.opt.locker.IsLocked(c.getLockKey(job.HandlerName)) {
		logger.LogYellow("cron_worker > job " + job.HandlerName + " is locked")
		return
	}
	defer c.opt.locker.Unlock(c.getLockKey(job.HandlerName))

	var err error
	trace, ctx := tracer.StartTraceFromHeader(ctx, "CronScheduler", map[string]string{})
	defer func() {
		if r := recover(); r != nil {
			trace.SetTag("panic", true)
			err = fmt.Errorf("%v", r)
		}
		logger.LogGreen("cron_scheduler > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish(tracer.FinishWithError(err))
	}()
	trace.SetTag("job_name", job.HandlerName)
	trace.Log("job_param", job.Params)

	if c.opt.debugMode {
		log.Printf("\x1b[35;3mCron Scheduler: executing task '%s' (interval: %s)\x1b[0m", job.HandlerName, job.Interval)
	}

	eventContext := candishared.NewEventContext(bytes.NewBuffer(make([]byte, 256)))
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(types.Scheduler))
	eventContext.SetHandlerRoute(job.HandlerName)
	eventContext.SetHeader(map[string]string{
		"interval": job.Interval,
	})
	eventContext.Write([]byte(job.Params))

	for _, handlerFunc := range job.Handler.HandlerFuncs {
		if err = handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
		}
	}
}

func (c *cronWorker) getLockKey(handlerName string) string {
	return fmt.Sprintf("%s:cron-worker-lock:%s", c.service.Name(), handlerName)
}

func (c *cronWorker) refreshWorker() {
	go func() { c.refreshWorkerNotif <- struct{}{} }()
}

func (c *cronWorker) stopAllJob() {
	for _, job := range c.activeJobs {
		job.ticker.Stop()
	}
}

func (c *cronWorker) registerNextInterval(j *Job) {

	if j.schedule != nil {
		j.ticker.Stop()
		j.ticker = time.NewTicker(j.schedule.NextInterval(time.Now()))
		c.workers[j.WorkerIndex].Chan = reflect.ValueOf(j.ticker.C)

	} else if j.nextDuration != nil {
		j.ticker.Stop()
		j.ticker = time.NewTicker(*j.nextDuration)
		c.workers[j.WorkerIndex].Chan = reflect.ValueOf(j.ticker.C)
		j.nextDuration = nil
	}

	c.refreshWorker()
}

// addJob to cron worker
func (c *cronWorker) addJob(job *Job) (err error) {

	if len(job.Handler.HandlerFuncs) == 0 {
		return errors.New("handler func cannot empty")
	}
	if job.HandlerName == "" {
		return errors.New("handler name cannot empty")
	}

	duration, nextDuration, err := candihelper.ParseDurationExpression(job.Interval)
	if err != nil {
		job.schedule, err = cronexpr.Parse(job.Interval)
		if err != nil {
			return err
		}
		duration = job.schedule.NextInterval(time.Now())
	}

	if nextDuration > 0 {
		job.nextDuration = &nextDuration
	}

	job.ticker = time.NewTicker(duration)
	job.WorkerIndex = len(c.workers)

	c.activeJobs = append(c.activeJobs, job)
	c.workers = append(c.workers, reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(job.ticker.C),
	})

	return nil
}
