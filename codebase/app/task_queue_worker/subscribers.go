package taskqueueworker

import (
	"context"
	"fmt"
	"math"
	"runtime"

	"github.com/golangid/candi/candihelper"
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
	if len(clientTaskJobListSubscribers) >= defaultOption.maxClientSubscriber {
		return errClientLimitExceeded
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientTaskJobListSubscribers[clientID] = clientTaskJobListSubscriber{
		c: clientChannel, filter: filter,
	}
	return nil
}

func removeJobListSubscriber(clientID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientTaskJobListSubscribers, clientID)
}

func registerNewJobDetailSubscriber(clientID, jobID string, clientChannel chan Job) error {
	if len(clientJobDetailSubscribers) >= defaultOption.maxClientSubscriber {
		return errClientLimitExceeded
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientJobDetailSubscribers[clientID] = clientJobDetailSubscriber{
		c: clientChannel, jobID: jobID,
	}
	return nil
}

func removeJobDetailSubscriber(clientID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientJobDetailSubscribers, clientID)
}

func broadcastAllToSubscribers(ctx context.Context) {
	if len(clientTaskSubscribers) > 0 {
		go broadcastTaskList(ctx)
	}
	if len(clientTaskJobListSubscribers) > 0 {
		go broadcastJobList(ctx)
	}
	if len(clientJobDetailSubscribers) > 0 {
		go broadcastJobDetail(ctx)
	}
}

func broadcastTaskList(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	var taskRes TaskListResolver
	taskRes.Data = persistent.AggregateAllTaskJob(ctx, Filter{TaskNameList: tasks})
	taskRes.Meta.TotalClientSubscriber = len(clientTaskSubscribers) + len(clientTaskJobListSubscribers) + len(clientJobDetailSubscribers)

	for _, subscriber := range clientTaskSubscribers {
		candihelper.TryCatch{
			Try: func() {
				subscriber <- taskRes
			},
			Catch: func(error) {},
		}.Do()
	}
}

func broadcastJobList(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	for _, subscriber := range clientTaskJobListSubscribers {
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
			Catch: func(error) {},
		}.Do()
	}
}

func broadcastJobDetail(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	for clientID, subscriber := range clientJobDetailSubscribers {
		detail, err := persistent.FindJobByID(ctx, subscriber.jobID)
		if err != nil || detail == nil {
			removeJobDetailSubscriber(clientID)
			continue
		}

		detail.updateValue()
		candihelper.TryCatch{
			Try: func() {
				subscriber.c <- *detail
			},
			Catch: func(error) {},
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
