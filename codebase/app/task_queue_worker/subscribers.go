package taskqueueworker

import (
	"context"
	"fmt"
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

	clientTaskSubscribers[clientID] = &clientTaskDashboardSubscriber{
		c: clientChannel, filter: filter,
	}
	return nil
}

func removeTaskListSubscriber(clientID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientTaskSubscribers, clientID)
}

func registerNewJobListSubscriber(clientID string, filter *Filter, clientChannel chan JobListResolver) error {
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

func registerNewJobDetailSubscriber(clientID string, filter *Filter, clientChannel chan JobResolver) error {
	if len(clientJobDetailSubscribers) >= defaultOption.maxClientSubscriber {
		return errClientLimitExceeded
	}

	mutex.Lock()
	defer mutex.Unlock()

	clientJobDetailSubscribers[clientID] = &clientJobDetailSubscriber{
		c: clientChannel, filter: filter,
	}
	return nil
}

func removeJobDetailSubscriber(clientID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientJobDetailSubscribers, clientID)
}

func broadcastAllToSubscribers(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	go func(ctx context.Context) {
		if len(clientTaskSubscribers) > 0 {
			broadcastTaskList(ctx)
		}
		if len(clientJobDetailSubscribers) > 0 {
			broadcastJobDetail(ctx)
		}
		if len(clientTaskJobListSubscribers) > 0 {
			broadcastJobList(ctx)
		}
	}(ctx)
}

func broadcastTaskList(ctx context.Context) {

	var taskRes TaskListResolver
	taskRes.Data = make([]TaskResolver, len(tasks))
	mapper := make(map[string]int, len(tasks))
	for i, task := range tasks {
		taskRes.Data[i].Name = task
		taskRes.Data[i].ModuleName = registeredTask[task].moduleName
		mapper[task] = i
	}

	for _, summary := range persistent.Summary().FindAllSummary(ctx, &Filter{}) {
		if idx, ok := mapper[summary.TaskName]; ok {
			res := TaskResolver{
				Name:       summary.TaskName,
				ModuleName: registeredTask[summary.TaskName].moduleName,
				TotalJobs:  summary.CountTotalJob(),
			}
			res.Detail = summary.ToSummaryDetail()
			res.IsLoading = summary.IsLoading
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

	subscriber, ok := clientTaskJobListSubscribers[clientID]
	if !ok {
		return
	}

	if subscriber.filter.TaskName != "" {
		summary := persistent.Summary().FindDetailSummary(ctx, subscriber.filter.TaskName)
		if summary.IsLoading {
			subscriber.skipBroadcast = summary.IsLoading
			subscriber.writeDataToChannel(JobListResolver{
				Meta: MetaJobList{IsLoading: summary.IsLoading},
			})
			return
		}
	}
	if subscriber.skipBroadcast {
		return
	}

	subscriber.filter.Sort = "-created_at"
	subscriber.skipBroadcast = candihelper.PtrToString(subscriber.filter.Search) != "" ||
		candihelper.PtrToString(subscriber.filter.JobID) != "" ||
		(!subscriber.filter.StartDate.IsZero() && !subscriber.filter.EndDate.IsZero())

	var jobListResolver JobListResolver
	jobListResolver.GetAllJob(ctx, subscriber.filter)
	jobListResolver.Meta.IsFreezeBroadcast = subscriber.skipBroadcast
	subscriber.writeDataToChannel(jobListResolver)
}

func broadcastJobDetail(ctx context.Context) {

	for clientID, subscriber := range clientJobDetailSubscribers {
		detail, err := persistent.FindJobByID(ctx, candihelper.PtrToString(subscriber.filter.JobID), subscriber.filter)
		if err != nil {
			removeJobDetailSubscriber(clientID)
			continue
		}
		var jobResolver JobResolver
		jobResolver.ParseFromJob(&detail)
		jobResolver.Meta.Page = subscriber.filter.Page
		jobResolver.Meta.TotalHistory = subscriber.filter.Count
		subscriber.writeDataToChannel(jobResolver)
	}
}

func broadcastWhenChangeAllJob(ctx context.Context, taskName string, isLoading bool) {

	persistent.Summary().UpdateSummary(ctx, taskName, map[string]interface{}{
		"is_loading": isLoading,
	})

	var taskRes TaskListResolver
	taskRes.Data = make([]TaskResolver, len(tasks))
	mapper := make(map[string]int, len(tasks))
	for i, task := range tasks {
		taskRes.Data[i].Name = task
		taskRes.Data[i].ModuleName = registeredTask[task].moduleName
		mapper[task] = i
	}

	for _, summary := range persistent.Summary().FindAllSummary(ctx, &Filter{}) {
		if idx, ok := mapper[summary.TaskName]; ok {
			res := TaskResolver{
				Name: summary.TaskName, ModuleName: registeredTask[summary.TaskName].moduleName,
				TotalJobs: summary.CountTotalJob(),
			}
			res.Detail = summary.ToSummaryDetail()
			res.IsLoading = summary.IsLoading
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

	for _, subscriber := range clientTaskJobListSubscribers {
		subscriber.skipBroadcast = isLoading
		subscriber.writeDataToChannel(JobListResolver{
			Meta: MetaJobList{IsLoading: isLoading},
		})
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
