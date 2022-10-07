package taskqueueworker

import (
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/golangid/candi/logger"
)

const (
	configurationRetentionAgeKey        = "retention_age"
	configurationClientSubscriberAgeKey = "client_subscriber_age"
	configurationMaxClientSubscriberKey = "max_client_subscriber"
	configurationTraceDetailURL         = "trace_detail_url"
)

type configurationUsecase struct {
	opt *option
}

func initConfiguration(opt *option) *configurationUsecase {
	defaultConfigs := []Configuration{
		{Key: configurationRetentionAgeKey, Name: "Retention Age", Value: "10m", IsActive: false},
		{Key: configurationClientSubscriberAgeKey, Name: "Client Subscriber Age", Value: "10m", IsActive: false},
		{Key: configurationMaxClientSubscriberKey, Name: "Max Client Subscriber", Value: "5", IsActive: false},
		{Key: configurationTraceDetailURL, Name: "Trace Detail URL", Value: "http://localhost:16686/trace", IsActive: true},
	}
	for _, cfg := range defaultConfigs {
		if _, err := opt.persistent.GetConfiguration(cfg.Key); err != nil {
			err := opt.persistent.SetConfiguration(&cfg)
			logger.LogIfError(err)
		}
	}

	return &configurationUsecase{
		opt: opt,
	}
}

func (c *configurationUsecase) getClientSubscriberAge() time.Duration {
	cfg, err := c.opt.persistent.GetConfiguration(configurationClientSubscriberAgeKey)
	if err != nil {
		return c.opt.autoRemoveClientInterval
	}
	age, err := time.ParseDuration(cfg.Value)
	if err != nil || !cfg.IsActive {
		return c.opt.autoRemoveClientInterval
	}
	return age
}

func (c *configurationUsecase) getMaxClientSubscriber() int {
	cfg, err := c.opt.persistent.GetConfiguration(configurationMaxClientSubscriberKey)
	if err != nil {
		return c.opt.maxClientSubscriber
	}
	max, err := strconv.Atoi(cfg.Value)
	if err != nil || !cfg.IsActive {
		return c.opt.maxClientSubscriber
	}
	return max
}

func (c *configurationUsecase) getTraceDetailURL() string {
	cfg, _ := c.opt.persistent.GetConfiguration(configurationTraceDetailURL)
	return cfg.Value
}

func (c *configurationUsecase) setConfiguration(cfg *Configuration) error {

	switch cfg.Key {
	case configurationRetentionAgeKey:

		interval, err := time.ParseDuration(cfg.Value)
		if err != nil || interval <= 0 {
			return errors.New("Invalid value")
		}

		taskIndex := engine.runningWorkerIndexTask[len(engine.workerChannels)-1]
		if taskIndex == nil {
			return errors.New("Missing task for worker")
		}
		if cfg.IsActive {
			taskIndex.activeInterval = time.NewTicker(interval)
			engine.workerChannels[len(engine.workerChannels)-1].Chan = reflect.ValueOf(taskIndex.activeInterval.C)
			engine.doRefreshWorker()
		} else if taskIndex.activeInterval != nil {
			taskIndex.activeInterval.Stop()
		}

	case configurationClientSubscriberAgeKey:
		interval, err := time.ParseDuration(cfg.Value)
		if err != nil || interval <= 0 {
			return errors.New("Invalid value")
		}

	case configurationMaxClientSubscriberKey:
		if val, err := strconv.Atoi(cfg.Value); err != nil {
			return err
		} else if val <= 0 {
			return errors.New("Must positive integer")
		}

	case configurationTraceDetailURL:

	default:
		return errors.New("Invalid config")
	}

	return c.opt.persistent.SetConfiguration(cfg)
}
