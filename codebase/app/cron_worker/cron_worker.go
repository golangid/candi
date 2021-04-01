package cronworker

// cron scheduler worker, create with 100% pure internal go library (using reflect select channel)

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"time"

	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candiutils"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

type cronWorker struct {
	service factory.ServiceFactory
	consul  *candiutils.Consul
	wg      sync.WaitGroup
}

// NewWorker create new cron worker
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	refreshWorkerNotif, shutdown = make(chan struct{}), make(chan struct{})
	semaphore = make(chan struct{}, env.BaseEnv().MaxGoroutines)
	startWorkerCh, releaseWorkerCh = make(chan struct{}), make(chan struct{})

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
				funcName, args, interval := candihelper.ParseCronJobKey(handler.Pattern)

				var job Job
				job.HandlerName = funcName
				job.HandlerFunc = handler.HandlerFunc
				job.Interval = interval
				job.Params = args
				if err := AddJob(job); err != nil {
					panic(fmt.Errorf(`Cron Worker: "%s" %v`, interval, err))
				}

				logger.LogYellow(fmt.Sprintf(`[CRON-WORKER] (job name): %s (every): %-8s  --> (module): "%s"`, `"`+funcName+`"`, interval, m.Name()))
			}
		}
	}
	fmt.Printf("\x1b[34;1m⇨ Cron worker running with %d jobs\x1b[0m\n\n", len(activeJobs))

	c := &cronWorker{
		service: service,
	}

	if env.BaseEnv().UseConsul {
		consul, err := candiutils.NewConsul(&candiutils.ConsulConfig{
			ConsulAgentHost:   env.BaseEnv().ConsulAgentHost,
			ConsulKey:         fmt.Sprintf("%s_cron_worker", service.Name()),
			LockRetryInterval: time.Second,
		})
		if err != nil {
			panic(err)
		}
		c.consul = consul
	}

	return c
}

func (c *cronWorker) Serve() {
	c.createConsulSession()

START:
	select {
	case <-startWorkerCh:
		startAllJob()
		totalRunJobs := 0

		// run worker
		for {
			chosen, _, ok := reflect.Select(workers)
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
			job := activeJobs[chosen]
			if job.nextDuration != nil {
				job.ticker.Stop()
				job.currentDuration = *job.nextDuration
				job.ticker = time.NewTicker(*job.nextDuration)
				workers[job.WorkerIndex].Chan = reflect.ValueOf(job.ticker.C)
				activeJobs[chosen].nextDuration = nil
			}

			semaphore <- struct{}{}
			c.wg.Add(1)
			go func(j *Job) {
				defer func() {
					c.wg.Done()
					<-semaphore
				}()

				c.processJob(j)
			}(job)

			if c.consul != nil {
				totalRunJobs++
				// if already running n jobs, release lock so that run in another instance
				if totalRunJobs == env.BaseEnv().ConsulMaxJobRebalance {
					// recreate session
					c.createConsulSession()
					<-releaseWorkerCh
					goto START
				}
			}
		}

	case <-shutdown:
		return
	}
}

func (c *cronWorker) Shutdown(ctx context.Context) {
	log.Println("\x1b[33;1mStopping Cron Job Scheduler worker...\x1b[0m")
	defer func() {
		if c.consul != nil {
			if err := c.consul.DestroySession(); err != nil {
				panic(err)
			}
		}
		log.Println("\x1b[33;1mStopping Cron Job Scheduler:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")
	}()

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

func (c *cronWorker) createConsulSession() {
	if c.consul == nil {
		go func() { startWorkerCh <- struct{}{} }()
		return
	}
	c.consul.DestroySession()
	stopAllJob()
	hostname, _ := os.Hostname()
	value := map[string]string{
		"hostname": hostname,
	}
	go c.consul.RetryLockAcquire(value, startWorkerCh, releaseWorkerCh)
}

func (c *cronWorker) processJob(job *Job) {
	trace := tracer.StartTrace(context.Background(), "CronScheduler")
	defer trace.Finish()
	ctx := trace.Context()

	defer func() {
		if r := recover(); r != nil {
			trace.SetError(fmt.Errorf("%v", r))
		}
		logger.LogGreen("cron scheduler > trace_url: " + tracer.GetTraceURL(ctx))
	}()

	if env.BaseEnv().DebugMode {
		log.Printf("\x1b[35;3mCron Scheduler: executing task '%s' (interval: %s)\x1b[0m", job.HandlerName, job.Interval)
	}

	tags := trace.Tags()
	tags["job_name"] = job.HandlerName
	if err := job.HandlerFunc(ctx, []byte(job.Params)); err != nil {
		trace.SetError(err)
	}
}
