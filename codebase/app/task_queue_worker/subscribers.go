package taskqueueworker

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sort"

	"github.com/golangid/candi/candihelper"
)

func registerNewTaskListSubscriber(clientID string, filter *Filter, clientChannel chan TaskListResolver) error {
	if len(clientTaskSubscribers) >= defaultOption.maxClientSubscriber {
		return errClientLimitExceeded
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientTaskSubscribers[clientID] = clientTaskDashboardSubscriber{
		c: clientChannel, filter: filter,
	}
	return nil
}

func removeTaskListSubscriber(clientID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientTaskSubscribers, clientID)
}

func registerNewJobListSubscriber(taskName, clientID string, filter *Filter, clientChannel chan JobListResolver) error {
	if len(clientTaskJobListSubscribers) >= defaultOption.maxClientSubscriber {
		return errClientLimitExceeded
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientTaskJobListSubscribers[clientID] = &clientTaskJobListSubscriber{
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

	taskSummary := persistent.FindAllSummary(ctx, &Filter{})

	var taskRes TaskListResolver
	taskRes.Data = make([]TaskResolver, len(tasks))
	mapper := make(map[string]int, len(tasks))
	for i, task := range tasks {
		taskRes.Data[i].Name = task
		taskRes.Data[i].ModuleName = registeredTask[task].moduleName
		mapper[task] = i
	}

	for _, summary := range taskSummary {
		if idx, ok := mapper[summary.TaskName]; ok {
			res := TaskResolver{
				Name:       summary.TaskName,
				ModuleName: registeredTask[summary.TaskName].moduleName,
				TotalJobs:  summary.CountTotalJob(),
			}
			res.Detail.Success = summary.Success
			res.Detail.Queueing = summary.Queueing
			res.Detail.Retrying = summary.Retrying
			res.Detail.Failure = summary.Failure
			res.Detail.Stopped = summary.Stopped
			taskRes.Data[idx] = res
		}
	}

	sort.Slice(taskRes.Data, func(i, j int) bool {
		return taskRes.Data[i].ModuleName < taskRes.Data[i].ModuleName
	})

	taskRes.Meta.TotalClientSubscriber = len(clientTaskSubscribers) + len(clientTaskJobListSubscribers) + len(clientJobDetailSubscribers)

	for _, subscriber := range clientTaskSubscribers {
		subscriber.writeDataToChannel(taskRes)
	}
}

func broadcastJobList(ctx context.Context) {
	for clientID := range clientTaskJobListSubscribers {
		broadcastJobListToClient(ctx, clientID)
	}
}

func broadcastJobListToClient(ctx context.Context, clientID string) {
	if ctx.Err() != nil {
		return
	}
	subscriber, ok := clientTaskJobListSubscribers[clientID]
	if ok {
		if subscriber.SkipBroadcast {
			return
		}
		subscriber.filter.Sort = "-created_at"
		jobs := persistent.FindAllJob(ctx, subscriber.filter)

		var meta MetaJobList
		subscriber.filter.TaskNameList = []string{subscriber.filter.TaskName}

		var taskDetailSummary []TaskSummary

		if candihelper.PtrToString(subscriber.filter.Search) != "" ||
			candihelper.PtrToString(subscriber.filter.JobID) != "" ||
			(!subscriber.filter.StartDate.IsZero() && !subscriber.filter.EndDate.IsZero()) {
			taskDetailSummary = persistent.AggregateAllTaskJob(ctx, subscriber.filter)
		} else {
			taskDetailSummary = persistent.FindAllSummary(ctx, subscriber.filter)
		}

		if len(taskDetailSummary) == 1 {
			meta.Detail = taskDetailSummary[0].ToSummaryDetail()
			meta.TotalRecords = taskDetailSummary[0].CountTotalJob()
		}
		meta.Page, meta.Limit = subscriber.filter.Page, subscriber.filter.Limit
		meta.TotalPages = int(math.Ceil(float64(meta.TotalRecords) / float64(meta.Limit)))

		subscriber.writeDataToChannel(JobListResolver{
			Meta: meta,
			Data: jobs,
		})
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
		subscriber.writeDataToChannel(*detail)
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
