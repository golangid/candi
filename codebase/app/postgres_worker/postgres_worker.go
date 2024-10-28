package postgresworker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/lib/pq"
)

/*
Postgres Event Listener Worker
Listen event from data change from selected table in postgres
*/

type (
	postgresWorker struct {
		ctx           context.Context
		ctxCancelFunc func()
		opt           option
		semaphores    map[string]chan struct{}
		shutdown      chan struct{}

		workerSourceIndex []string
		workers           []reflect.SelectCase
		service           factory.ServiceFactory
		wg                sync.WaitGroup

		messagePool sync.Pool
	}
)

// NewWorker create new postgres event listener
func NewWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	worker := &postgresWorker{
		service:    service,
		opt:        getDefaultOption(service),
		semaphores: make(map[string]chan struct{}),
		shutdown:   make(chan struct{}, 1),
		messagePool: sync.Pool{
			New: func() any {
				return candishared.NewEventContext(bytes.NewBuffer(make([]byte, 0, 256)))
			},
		},
	}

	for _, opt := range opts {
		opt(&worker.opt)
	}

	if len(worker.opt.sources) == 0 {
		worker.opt.sources[""] = &PostgresSource{dsn: env.BaseEnv().DbSQLWriteDSN} // default source
	}

	for _, source := range worker.opt.sources {
		source.db, source.listener = getListener(source.dsn, &worker.opt)
		source.handlers = make(map[string]types.WorkerHandler)
		source.workerIndex = len(worker.workerSourceIndex)

		worker.workerSourceIndex = append(worker.workerSourceIndex, source.name)
		worker.workers = append(worker.workers, reflect.SelectCase{
			Dir: reflect.SelectRecv, Chan: reflect.ValueOf(source.listener.Notify),
		})

		if err := source.execCreateFunctionEventQuery(); err != nil {
			log.Panicf("[POSTGRES-LISTENER]%s: failed when create event function: %s%s",
				getWorkerTypeLog(worker.opt.workerType), err, source.getLogForSourceName())
		}
	}

	worker.opt.locker.Reset(fmt.Sprintf("%s:postgres-worker-lock:*", service.Name()))
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(worker.opt.workerType); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				sourceName, tableName := ParseHandlerRoute(handler.Pattern)
				postgresSource, ok := worker.opt.sources[sourceName]
				if !ok || postgresSource == nil {
					log.Panicf("[POSTGRES-LISTENER]%s: source name '%s' unregistered (when register table '%s' in module '%s')",
						getWorkerTypeLog(worker.opt.workerType), sourceName, tableName, m.Name())
				}

				if err := postgresSource.execTriggerQuery(tableName); err != nil {
					log.Panicf("[POSTGRES-LISTENER]%s: failed when create trigger for table %s%s: %s",
						getWorkerTypeLog(worker.opt.workerType), tableName, postgresSource.getLogForSourceName(), err)
				}

				postgresSource.handlers[tableName] = handler

				worker.semaphores[tableName] = make(chan struct{}, worker.opt.maxGoroutines)
				worker.opt.sources[sourceName] = postgresSource
				logger.LogYellow(fmt.Sprintf(`[POSTGRES-LISTENER]%s (table): "%s"%s  --> (module): "%s"`,
					getWorkerTypeLog(worker.opt.workerType), tableName, postgresSource.getLogForSourceName(), m.Name()))
			}
		}
	}

	if len(worker.semaphores) == 0 {
		log.Printf("postgres listener%s: no table event provided", getWorkerTypeLog(worker.opt.workerType))
	} else {
		fmt.Printf("\x1b[34;1m⇨ Postgres Event Listener%s running with %d handlers\x1b[0m\n\n", getWorkerTypeLog(worker.opt.workerType), len(worker.semaphores))
	}

	worker.ctx, worker.ctxCancelFunc = context.WithCancel(context.Background())
	return worker
}

func (p *postgresWorker) Serve() {
	for _, source := range p.opt.sources {
		source.listener.Listen(eventsConst)
	}

	// run worker
	for {
		select {
		case <-p.shutdown:
			return

		default:
		}

		chosen, value, ok := reflect.Select(p.workers)
		if !ok {
			continue
		}

		// exec handler
		if e, ok := value.Interface().(*pq.Notification); ok && e != nil {
			var payload EventPayload
			json.Unmarshal([]byte(e.Extra), &payload)

			p.semaphores[payload.Table] <- struct{}{}
			p.wg.Add(1)
			go func(data *EventPayload, workerIndex int) {
				defer func() { p.wg.Done(); <-p.semaphores[data.Table] }()

				if p.ctx.Err() != nil {
					logger.LogRed("postgres_listener > ctx root err: " + p.ctx.Err().Error())
					return
				}

				p.execEvent(workerIndex, data)
			}(&payload, chosen)
		}
	}
}

func (p *postgresWorker) Shutdown(ctx context.Context) {
	defer func() {
		fmt.Printf("\r%s \x1b[33;1mPostgres Event Listener%s:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m%s\n",
			time.Now().Format(candihelper.TimeFormatLogger), getWorkerTypeLog(p.opt.workerType), strings.Repeat(" ", 20))
	}()

	p.shutdown <- struct{}{}
	runningJob := 0
	for _, sem := range p.semaphores {
		runningJob += len(sem)
	}
	waitingJob := "... "
	if runningJob != 0 {
		waitingJob = fmt.Sprintf("waiting %d job until done... ", runningJob)
	}
	fmt.Printf("\r%s \x1b[33;1mPostgres Event Listener%s:\x1b[0m %s",
		time.Now().Format(candihelper.TimeFormatLogger), getWorkerTypeLog(p.opt.workerType), waitingJob)

	for _, source := range p.opt.sources {
		source.listener.Close()
	}
	p.wg.Wait()
	p.ctxCancelFunc()
	p.opt.locker.Reset(fmt.Sprintf("%s:postgres-worker-lock:*", p.service.Name()))
}

func (p *postgresWorker) Name() string {
	return string(p.opt.workerType)
}

func (p *postgresWorker) getLockKey(eventPayload *EventPayload) string {
	return fmt.Sprintf("%s:postgres-worker-lock:%s-%s-%s", p.service.Name(), eventPayload.Table, eventPayload.Action, eventPayload.EventID)
}

func (p *postgresWorker) execEvent(workerIndex int, data *EventPayload) {
	source, ok := p.opt.sources[p.workerSourceIndex[workerIndex]]
	if !ok || source == nil {
		return
	}

	// lock for multiple worker (if running on multiple pods/instance)
	if p.opt.locker.IsLocked(p.getLockKey(data)) {
		return
	}
	defer p.opt.locker.Unlock(p.getLockKey(data))

	ctx := p.ctx
	handler, ok := source.handlers[data.Table]
	if !ok {
		return
	}

	if handler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	var err error
	trace, ctx := tracer.StartTraceFromHeader(ctx, "PostgresEventListener", make(map[string]string, 0))
	defer func() {
		if r := recover(); r != nil {
			trace.SetTag("panic", true)
			err = fmt.Errorf("%v", r)
		}
		trace.Finish(tracer.FinishWithError(err))
	}()

	if p.opt.debugMode {
		var sourceLog string
		if source.name != "" {
			sourceLog = fmt.Sprintf(" (from source name '%s')", source.name)
		}
		log.Printf("\x1b[35;3mPostgres Event Listener%s: executing event from table: '%s'%s and action: '%s'\x1b[0m", getWorkerTypeLog(p.opt.workerType), data.Table, sourceLog, data.Action)
	}

	if data.Data.IsTooLongPayload {
		detailData := source.findDetailData(data.Table, data.GetID())
		switch data.Action {
		case ActionInsert:
			data.Data.New = detailData
		case ActionUpdate:
			data.Data.New = detailData
			data.Data.Old = detailData
		case ActionDelete:
			data.Data.Old = detailData
		}
	}

	eventContext := p.messagePool.Get().(*candishared.EventContext)
	defer p.releaseMessagePool(eventContext)
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(p.opt.workerType))
	eventContext.SetHandlerRoute(data.Table)
	eventContext.SetKey(data.EventID)

	message, _ := json.Marshal(data)
	eventContext.Write(message)

	if source.name != "" {
		trace.SetTag("source_name", source.name)
		eventContext.SetHeader(map[string]string{
			"source_name": source.name,
		})
	}
	if p.opt.workerType != types.PostgresListener {
		trace.SetTag("worker_type", string(p.opt.workerType))
	}
	trace.SetTag("table_name", data.Table)
	trace.SetTag("action", data.Action)
	trace.Log("dsn", candihelper.MaskingPasswordURL(source.dsn))
	trace.Log("payload", data)

	for _, handlerFunc := range handler.HandlerFuncs {
		if err = handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
		}
	}
}

func (p *postgresWorker) releaseMessagePool(eventContext *candishared.EventContext) {
	eventContext.Reset()
	p.messagePool.Put(eventContext)
}

func getWorkerTypeLog(name types.Worker) (workerType string) {
	if name != types.PostgresListener {
		workerType = " [worker_type: " + string(name) + "]"
	}
	return
}
