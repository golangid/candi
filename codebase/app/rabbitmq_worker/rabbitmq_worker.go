package rabbitmqworker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/golangid/candi/broker"
	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitmqWorker struct {
	ctx           context.Context
	ctxCancelFunc func()
	opt           option

	bk *broker.RabbitMQBroker

	shutdown   chan struct{}
	isShutdown bool
	semaphore  []chan struct{}
	wg         sync.WaitGroup
	receiver   []reflect.SelectCase
	handlers   map[string]types.WorkerHandler
}

// NewWorker create new rabbitmq consumer
func NewWorker(service factory.ServiceFactory, bk interfaces.Broker, opts ...OptionFunc) factory.AppServerFactory {
	rabbitMQBroker, ok := bk.(*broker.RabbitMQBroker)
	if !ok {
		panic("Missing RabbitMQ broker configuration")
	}

	worker := &rabbitmqWorker{
		opt: getDefaultOption(),
		bk:  rabbitMQBroker,
	}
	for _, opt := range opts {
		opt(&worker.opt)
	}

	worker.ctx, worker.ctxCancelFunc = context.WithCancel(context.Background())

	worker.shutdown = make(chan struct{}, 1)
	worker.handlers = make(map[string]types.WorkerHandler)

	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(rabbitMQBroker.WorkerType); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				logger.LogYellow(fmt.Sprintf(`[RABBITMQ-CONSUMER]%s (queue): %-15s  --> (module): "%s"`, getWorkerTypeLog(rabbitMQBroker.WorkerType), `"`+handler.Pattern+`"`, m.Name()))
				queueChan, err := setupQueueConfig(worker.bk.Channel, worker.opt.consumerGroup, rabbitMQBroker.Exchange, handler.Pattern)
				if err != nil {
					panic(err)
				}

				worker.receiver = append(worker.receiver, reflect.SelectCase{
					Dir: reflect.SelectRecv, Chan: reflect.ValueOf(queueChan),
				})
				worker.handlers[handler.Pattern] = handler
				worker.semaphore = append(worker.semaphore, make(chan struct{}, 1))
			}
		}
	}

	fmt.Printf("\x1b[34;1m⇨ RabbitMQ consumer%s running with %d queue. Broker: %s\x1b[0m\n\n", getWorkerTypeLog(rabbitMQBroker.WorkerType), len(worker.receiver),
		candihelper.MaskingPasswordURL(rabbitMQBroker.BrokerHost))

	return worker
}

func (r *rabbitmqWorker) Serve() {
	for {
		select {
		case <-r.shutdown:
			return

		default:
		}

		chosen, value, ok := reflect.Select(r.receiver)
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
				defer func() {
					r.wg.Done()
					<-r.semaphore[idx]
				}()
				r.processMessage(message)
			}(msg, chosen)
		}
	}
}

func (r *rabbitmqWorker) Shutdown(ctx context.Context) {
	defer log.Printf("\x1b[33;1mStopping RabbitMQ Worker%s:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m\n", getWorkerTypeLog(r.bk.WorkerType))

	r.shutdown <- struct{}{}
	r.isShutdown = true
	var runningJob int
	for _, sem := range r.semaphore {
		runningJob += len(sem)
	}
	if runningJob != 0 {
		fmt.Printf("\x1b[34;1mRabbitMQ Worker%s:\x1b[0m waiting %d job until done...\x1b[0m\n", getWorkerTypeLog(r.bk.WorkerType), runningJob)
	}

	r.wg.Wait()
	r.bk.Channel.Close()
	r.ctxCancelFunc()
}

func (r *rabbitmqWorker) Name() string {
	return string(r.bk.WorkerType)
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

	header := make(map[string]string, len(message.Headers))
	for key, val := range message.Headers {
		header[key] = string(candihelper.ToBytes(val))
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "RabbitMQConsumer", header)
	defer trace.Finish(
		tracer.FinishWithRecoverPanic(func(any) {}),
		tracer.FinishWithFunc(func() {
			if selectedHandler.AutoACK {
				message.Ack(false)
			}
		}),
	)

	trace.SetTag("broker", candihelper.MaskingPasswordURL(r.bk.BrokerHost))
	trace.SetTag("exchange", message.Exchange)
	trace.SetTag("routing_key", message.RoutingKey)
	if r.bk.WorkerType != types.RabbitMQ {
		trace.SetTag("worker_type", string(r.bk.WorkerType))
	}
	trace.Log("header", message.Headers)
	trace.Log("body", message.Body)

	if r.opt.debugMode {
		log.Printf("\x1b[35;3mRabbitMQ Consumer%s: message consumed, topic = %s\x1b[0m", getWorkerTypeLog(r.bk.WorkerType), message.RoutingKey)
	}

	eventContext := candishared.NewEventContext(bytes.NewBuffer(make([]byte, 256)))
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(r.bk.WorkerType))
	eventContext.SetHandlerRoute(message.RoutingKey)
	eventContext.SetHeader(header)
	eventContext.SetKey(message.Exchange)
	eventContext.Write(message.Body)

	for _, handlerFunc := range selectedHandler.HandlerFuncs {
		if err := handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
			trace.SetError(err)
		}
	}
}

func getWorkerTypeLog(name types.Worker) (workerType string) {
	if name != types.RabbitMQ {
		workerType = " [worker_type: " + string(name) + "]"
	}
	return
}
