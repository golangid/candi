package taskqueueworker

import (
	"context"
	"fmt"
	"math"
	"runtime"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
)

func registerNewTaskListSubscriber(clientID string, clientChannel chan TaskListResolver) error {
	if len(clientTaskSubscribers) >= defaultOption.maxClientSubscriber {
		return errClientLimitExceeded
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientTaskSubscribers[clientID] = clientChannel
	return nil
}

func removeTaskListSubscriber(clientID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientTaskSubscribers, clientID)
}

func registerNewJobListSubscriber(taskName, clientID string, filter Filter, clientChannel chan JobListResolver) error {
	if len(clientJobTaskSubscribers) >= defaultOption.maxClientSubscriber {
		return errClientLimitExceeded
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientJobTaskSubscribers[clientID] = clientJobTaskSubscriber{
		c: clientChannel, filter: filter,
	}
	return nil
}

func removeJobListSubscriber(taskName, clientID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientJobTaskSubscribers, clientID)
}

func broadcastAllToSubscribers(ctx context.Context) {
	if len(clientTaskSubscribers) > 0 {
		go broadcastTaskList(ctx)
	}
	if len(clientJobTaskSubscribers) > 0 {
		go broadcastJobList(ctx)
	}
}

func broadcastTaskList(ctx context.Context) {
	if ctx.Err() != nil {
		logger.LogI(ctx.Err().Error())
		return
	}

	var taskRes TaskListResolver
	taskRes.Data = persistent.AggregateAllTaskJob(ctx, Filter{TaskNameList: tasks})
	taskRes.Meta.TotalClientSubscriber = len(clientTaskSubscribers) + len(clientJobTaskSubscribers)

	for _, subscriber := range clientTaskSubscribers {
		candihelper.TryCatch{
			Try: func() {
				subscriber <- taskRes
			},
			Catch: func(e error) {
				logger.LogE(e.Error())
			},
		}.Do()
	}
}

func broadcastJobList(ctx context.Context) {
	if ctx.Err() != nil {
		logger.LogI(ctx.Err().Error())
		return
	}
	for _, subscriber := range clientJobTaskSubscribers {
		jobs := persistent.FindAllJob(ctx, subscriber.filter)

		var meta MetaJobList
		subscriber.filter.TaskNameList = []string{subscriber.filter.TaskName}
		counterAll := persistent.AggregateAllTaskJob(ctx, subscriber.filter)
		if len(counterAll) == 1 {
			meta.Detail.Failure = counterAll[0].Detail.Failure
			meta.Detail.Retrying = counterAll[0].Detail.Retrying
			meta.Detail.Success = counterAll[0].Detail.Success
			meta.Detail.Queueing = counterAll[0].Detail.Queueing
			meta.Detail.Stopped = counterAll[0].Detail.Stopped
			meta.TotalRecords = counterAll[0].TotalJobs
		}
		meta.Page, meta.Limit = subscriber.filter.Page, subscriber.filter.Limit
		meta.TotalPages = int(math.Ceil(float64(meta.TotalRecords) / float64(meta.Limit)))

		candihelper.TryCatch{
			Try: func() {
				subscriber.c <- JobListResolver{
					Meta: meta,
					Data: jobs,
				}
			},
			Catch: func(e error) {
				logger.LogE(e.Error())
			},
		}.Do()
	}
}

func getMemstats() (res MemstatsResolver) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	res.Alloc = fmt.Sprintf("%d MB", m.Alloc/candihelper.MByte)

	if m.TotalAlloc > candihelper.GByte {
		res.TotalAlloc = fmt.Sprintf("%.2f GB", float64(m.TotalAlloc)/float64(candihelper.GByte))
	} else {
		res.TotalAlloc = fmt.Sprintf("%d MB", m.TotalAlloc/candihelper.MByte)
	}
	res.NumGC = int(m.NumGC)
	res.NumGoroutines = runtime.NumGoroutine()
	return
}
