package taskqueueworker

import (
	"reflect"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	cronexpr "github.com/golangid/candi/candiutils/cronparser"
	"github.com/golangid/candi/logger"
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
		var err error
		internalTaskRetention.schedule, err = cronexpr.Parse(cfg.Value)
		if err != nil {
			goto END
		}
		internalTaskRetention.activeInterval = time.NewTicker(internalTaskRetention.schedule.NextInterval(time.Now()))
		retentionBeat.Chan = reflect.ValueOf(internalTaskRetention.activeInterval.C)
	}

END:
	t.runningWorkerIndexTask[internalTaskRetention.workerIndex] = internalTaskRetention
	t.workerChannels = append(t.workerChannels, retentionBeat)

}

func (t *taskQueueWorker) execInternalTask(task *Task) {

	logger.LogIf("running internal task: %s", task.internalTaskName)

	switch task.internalTaskName {
	case configurationRetentionAgeKey:

		cfg, _ := t.opt.persistent.GetConfiguration(configurationRetentionAgeKey)
		if !cfg.IsActive {
			return
		}

		if task.schedule == nil {
			return
		}

		now := time.Now()
		interval := task.schedule.NextInterval(now)
		task.activeInterval = time.NewTicker(interval)
		t.workerChannels[task.workerIndex].Chan = reflect.ValueOf(task.activeInterval.C)
		t.doRefreshWorker()

		lockKey := t.getLockKey("internal_task:" + task.internalTaskName)
		if t.opt.locker.IsLocked(lockKey) {
			logger.LogI("task_queue_worker > internal task " + task.internalTaskName + " is locked")
			return
		}
		defer t.opt.locker.Unlock(lockKey)

		beforeCreatedAt := now.Add(-interval)

		// only remove success job
		for _, task := range t.tasks {
			incrQuery := map[string]int64{}
			countAffected := t.opt.persistent.CleanJob(t.ctx,
				&Filter{
					TaskName: task, BeforeCreatedAt: &beforeCreatedAt,
					Status: candihelper.ToStringPtr(string(StatusSuccess)),
				},
			)
			incrQuery[strings.ToLower(string(StatusSuccess))] -= countAffected
			t.opt.persistent.Summary().IncrementSummary(t.ctx, task, incrQuery)
		}
		t.subscriber.broadcastAllToSubscribers(t.ctx)
	}

}
