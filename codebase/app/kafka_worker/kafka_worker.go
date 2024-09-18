package kafkaworker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/golangid/candi/broker"
	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/logger"
)

type kafkaWorker struct {
	bk              *broker.KafkaBroker
	opt             option
	engine          sarama.ConsumerGroup
	service         factory.ServiceFactory
	consumerHandler *consumerHandler
	cancelFunc      func()
}

// NewWorker create new kafka consumer
func NewWorker(service factory.ServiceFactory, bk interfaces.Broker, opts ...OptionFunc) factory.AppServerFactory {
	kafkaBroker, ok := bk.(*broker.KafkaBroker)
	if !ok {
		panic("Missing Kafka broker configuration")
	}

	worker := &kafkaWorker{
		bk:      kafkaBroker,
		service: service,
		opt:     getDefaultOption(),
	}
	for _, opt := range opts {
		opt(&worker.opt)
	}

	// init kafka consumer
	consumerEngine, err := sarama.NewConsumerGroupFromClient(
		worker.opt.consumerGroup,
		kafkaBroker.Client,
	)
	if err != nil {
		log.Panicf("Error creating kafka consumer group client: %v", err)
	}

	var consumerHandler consumerHandler
	consumerHandler.bk = kafkaBroker
	consumerHandler.handlerFuncs = make(map[string]types.WorkerHandler)
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(worker.bk.WorkerType); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				if _, ok := consumerHandler.handlerFuncs[handler.Pattern]; ok {
					logger.LogYellow(fmt.Sprintf("Kafka: warning, topic %s has been used in another module, overwrite handler func", handler.Pattern))
				}
				consumerHandler.handlerFuncs[handler.Pattern] = handler
				consumerHandler.topics = append(consumerHandler.topics, handler.Pattern)
				logger.LogYellow(fmt.Sprintf(`[KAFKA-CONSUMER]%s (topic): %-15s  --> (module): "%s"`, getWorkerTypeLog(kafkaBroker.WorkerType), `"`+handler.Pattern+`"`, m.Name()))
			}
		}
	}
	fmt.Printf("\x1b[34;1m⇨ Kafka consumer%s running with %d topics. Brokers: "+strings.Join(kafkaBroker.BrokerHost, ", ")+"\x1b[0m\n\n",
		getWorkerTypeLog(kafkaBroker.WorkerType), len(consumerHandler.topics))

	consumerHandler.ready = make(chan struct{})
	consumerHandler.opt = &worker.opt
	consumerHandler.messagePool = sync.Pool{
		New: func() interface{} {
			return candishared.NewEventContext(bytes.NewBuffer(make([]byte, 0, 256)))
		},
	}

	worker.engine = consumerEngine
	worker.consumerHandler = &consumerHandler
	return worker
}

func (h *kafkaWorker) Serve() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancelFunc = cancel

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := h.engine.Consume(ctx, h.consumerHandler.topics, h.consumerHandler); err != nil {
				logger.LogRed(fmt.Sprintf("Error from kafka consumer%s: %s", getWorkerTypeLog(h.bk.WorkerType), err.Error()))
				if errCode, ok := err.(sarama.KError); ok {
					switch errCode {
					case sarama.ErrInvalidTopic:
						log.Fatal(errCode.Error())
					}
				}
			}

			if ctx.Err() != nil {
				return
			}
			h.consumerHandler.ready = make(chan struct{})
		}
	}()

	<-h.consumerHandler.ready
	wg.Wait()
}

func (h *kafkaWorker) Shutdown(ctx context.Context) {
	defer func() {
		fmt.Printf("\r%s \x1b[33;1mStopping Kafka Consumer%s:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m%s\n",
			time.Now().Format(candihelper.TimeFormatLogger), getWorkerTypeLog(h.bk.WorkerType), strings.Repeat(" ", 20))
	}()

	fmt.Printf("\r%s \x1b[33;1mStopping Kafka Consumer%s:\x1b[0m ... ", time.Now().Format(candihelper.TimeFormatLogger), getWorkerTypeLog(h.bk.WorkerType))
	h.cancelFunc()
	h.engine.Close()
}

func (h *kafkaWorker) Name() string {
	return string(h.bk.WorkerType)
}

func getWorkerTypeLog(name types.Worker) (workerType string) {
	if name != types.Kafka {
		workerType = " [worker_type: " + string(name) + "]"
	}
	return
}
