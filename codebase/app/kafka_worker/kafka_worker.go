package kafkaworker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
)

type kafkaWorker struct {
	opt             option
	engine          sarama.ConsumerGroup
	service         factory.ServiceFactory
	consumerHandler *consumerHandler
	cancelFunc      func()
}

// NewWorker create new kafka consumer
func NewWorker(service factory.ServiceFactory, opts ...OptionFunc) factory.AppServerFactory {
	worker := &kafkaWorker{
		service: service,
		opt:     getDefaultOption(),
	}

	for _, opt := range opts {
		opt(&worker.opt)
	}

	// init kafka consumer
	if service.GetDependency().GetBroker(types.Kafka) == nil {
		log.Panic("Missing Kafka configuration")
	}
	consumerEngine, err := sarama.NewConsumerGroupFromClient(
		worker.opt.consumerGroup,
		service.GetDependency().GetBroker(types.Kafka).GetConfiguration().(sarama.Client),
	)
	if err != nil {
		log.Panicf("Error creating kafka consumer group client: %v", err)
	}

	var consumerHandler consumerHandler
	consumerHandler.handlerFuncs = make(map[string]types.WorkerHandler)
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.Kafka); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				if _, ok := consumerHandler.handlerFuncs[handler.Pattern]; ok {
					logger.LogYellow(fmt.Sprintf("Kafka: warning, topic %s has been used in another module, overwrite handler func", handler.Pattern))
				}
				consumerHandler.handlerFuncs[handler.Pattern] = handler
				consumerHandler.topics = append(consumerHandler.topics, handler.Pattern)
				logger.LogYellow(fmt.Sprintf(`[KAFKA-CONSUMER] (topic): %-15s  --> (module): "%s"`, `"`+handler.Pattern+`"`, m.Name()))
			}
		}
	}
	fmt.Printf("\x1b[34;1m⇨ Kafka consumer running with %d topics. Brokers: "+strings.Join(worker.opt.brokers, ", ")+"\x1b[0m\n\n",
		len(consumerHandler.topics))

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
				logger.LogRed("Error from kafka consumer: " + err.Error())
				if errCode, ok := err.(sarama.KError); ok {
					switch errCode {
					case sarama.ErrInvalidTopic:
						log.Fatal(errCode.Error())
					}
				}
				time.Sleep(time.Second)
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
	defer log.Println("\x1b[33;1mStopping Kafka Consumer:\x1b[0m \x1b[32;1mSUCCESS\x1b[0m")

	h.cancelFunc()
	h.engine.Close()
}

func (h *kafkaWorker) Name() string {
	return string(types.Kafka)
}
