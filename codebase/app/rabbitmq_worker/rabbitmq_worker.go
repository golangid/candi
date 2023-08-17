package rabbitmqworker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/streadway/amqp"
)

type rabbitmqWorker struct {
	ctx           context.Context
	ctxCancelFunc func()
	opt           option

	ch         *amqp.Channel
	shutdown   chan struct{}
	isShutdown bool
	semaphore  []chan struct{}
	wg         sync.WaitGroup
	channels   []reflect.SelectCase
	handlers   map[string]types.WorkerHandler
}

// NewWorker create new rabbitmq consumer
func NewWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	if service.GetDependency().GetBroker(types.RabbitMQ) == nil {
		panic("Missing RabbitMQ configuration")
	}

	worker := &rabbitmqWorker{
		opt: getDefaultOption(),
	}
	for _, opt := range opts {
		opt(&worker.opt)
	}

	worker.ctx, worker.ctxCancelFunc = context.WithCancel(context.Background())
	worker.ch = service.GetDependency().GetBroker(types.RabbitMQ).GetConfiguration().(*amqp.Channel)

	worker.shutdown = make(chan struct{}, 1)
	worker.handlers = make(map[string]types.WorkerHandler)

	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.RabbitMQ); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				logger.LogYellow(fmt.Sprintf(`[RABBITMQ-CONSUMER] (queue): %-15s  --> (module): "%s"`, `"`+handler.Pattern+`"`, m.Name()))
				queueChan, err := setupQueueConfig(worker.ch, worker.opt.consumerGroup, worker.opt.exchangeName, handler.Pattern)
				if err != nil {
					panic(err)
				}

				worker.channels = append(worker.channels, reflect.SelectCase{
					Dir: reflect.SelectRecv, Chan: reflect.ValueOf(queueChan),
				})
				worker.handlers[handler.Pattern] = handler
				worker.semaphore = append(worker.semaphore, make(chan struct{}, 1))
			}
		}
	}

	fmt.Printf("\x1b[34;1m⇨ RabbitMQ consumer running with %d queue. Broker: %s\x1b[0m\n\n", len(worker.channels),
		candihelper.MaskingPasswordURL(worker.opt.broker))

	return worker
}

func (r *rabbitmqWorker) Serve() {

	for {
		select {
		case <-r.shutdown:
			return

		default:
		}

		chosen, value, ok := reflect.Select(r.channels)
		if !ok {
			continue
		}

		// exec handler
		if msg, ok := value.Interface().(amqp.Delivery); ok {
			r.semaphore[chosen] <- struct{}{}
			if r.isShutdown {
				return
			}

			r.wg.Add(1)
			go func(message amqp.Delivery, idx int) {
				r.processMessage(message)
				r.wg.Done()
				<-r.semaphore[idx]
			}(msg, chosen)
		}
	}
}

func (r *rabbitmqWorker) Shutdown(ctx context.Context) {
	defer log.Println("\x1b[33;1mStopping RabbitMQ Worker:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	r.shutdown <- struct{}{}
	r.isShutdown = true
	var runningJob int
	for _, sem := range r.semaphore {
		runningJob += len(sem)
	}
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mRabbitMQ Worker:\x1b[0m waiting %d job until done...\x1b[0m\n", runningJob)
	}

	r.wg.Wait()
	r.ch.Close()
	r.ctxCancelFunc()
}

func (r *rabbitmqWorker) Name() string {
	return string(types.RabbitMQ)
}

func (r *rabbitmqWorker) processMessage(message amqp.Delivery) {
	if r.ctx.Err() != nil {
		logger.LogRed("rabbitmq_consumer > ctx root err: " + r.ctx.Err().Error())
		return
	}

	ctx := r.ctx
	selectedHandler := r.handlers[message.RoutingKey]
	if selectedHandler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	header := map[string]string{}
	for key, val := range message.Headers {
		header[key] = string(candihelper.ToBytes(val))
	}

	var err error
	trace, ctx := tracer.StartTraceFromHeader(ctx, "RabbitMQConsumer", header)
	defer func() {
		if r := recover(); r != nil {
			trace.SetTag("panic", true)
			err = fmt.Errorf("%v", r)
		}
		if selectedHandler.AutoACK {
			message.Ack(false)
		}
		logger.LogGreen("rabbitmq_consumer > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("broker", candihelper.MaskingPasswordURL(r.opt.broker))
	trace.SetTag("exchange", message.Exchange)
	trace.SetTag("routing_key", message.RoutingKey)
	trace.Log("header", message.Headers)
	trace.Log("body", message.Body)

	if r.opt.debugMode {
		log.Printf("\x1b[35;3mRabbitMQ Consumer: message consumed, topic = %s\x1b[0m", message.RoutingKey)
	}

	eventContext := candishared.NewEventContext(bytes.NewBuffer(make([]byte, 256)))
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(types.RabbitMQ))
	eventContext.SetHandlerRoute(message.RoutingKey)
	eventContext.SetHeader(header)
	eventContext.SetKey(message.Exchange)
	eventContext.Write(message.Body)

	for _, handlerFunc := range selectedHandler.HandlerFuncs {
		if err = handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
		}
	}
}
