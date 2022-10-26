package taskqueueworker

import (
	"reflect"
	"time"

	"github.com/golangid/candi/candihelper"
)

func (t *taskQueueWorker) registerInternalTask() {

	retentionBeat := reflect.SelectCase{Dir: reflect.SelectRecv}
	internalTaskRetention := &Task{
		isInternalTask:   true,
		internalTaskName: configurationRetentionAgeKey,
		workerIndex:      len(t.workerChannels),
	}
	cfg, _ := t.opt.persistent.GetConfiguration(configurationRetentionAgeKey)
	if cfg.IsActive {
		interval, nextInterval, err := candihelper.ParseAtTime(cfg.Value)
		if err != nil {
			return
		}
		if nextInterval > 0 {
			internalTaskRetention.nextInterval = &nextInterval
		}
		internalTaskRetention.activeInterval = time.NewTicker(interval)
		retentionBeat.Chan = reflect.ValueOf(internalTaskRetention.activeInterval.C)
	}
	t.runningWorkerIndexTask[internalTaskRetention.workerIndex] = internalTaskRetention
	t.workerChannels = append(t.workerChannels, retentionBeat)

}
