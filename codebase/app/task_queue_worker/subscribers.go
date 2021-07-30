package taskqueueworker

import (
	"fmt"
	"runtime"

	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/logger"
)

func registerNewTaskListSubscriber(clientID string, clientChannel chan TaskListResolver) error {
	if len(clientTaskSubscribers) >= defaultOption.MaxClientSubscriber {
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
	if len(clientJobTaskSubscribers) >= defaultOption.MaxClientSubscriber {
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

func broadcastAllToSubscribers() {
	if len(clientTaskSubscribers) > 0 {
		go broadcastTaskList()
	}
	if len(clientJobTaskSubscribers) > 0 {
		go broadcastJobList()
	}
}

func broadcastTaskList() {
	var taskRes TaskListResolver
	for _, task := range tasks {
		var tsk = TaskResolver{
			Name: task,
		}
		tsk.Detail.GiveUp = repo.countTaskJobDetail(task, statusFailure)
		tsk.Detail.Retrying = repo.countTaskJobDetail(task, statusRetrying)
		tsk.Detail.Success = repo.countTaskJobDetail(task, statusSuccess)
		tsk.Detail.Queueing = repo.countTaskJobDetail(task, statusQueueing)
		tsk.Detail.Stopped = repo.countTaskJobDetail(task, statusStopped)
		tsk.TotalJobs = tsk.Detail.GiveUp + tsk.Detail.Retrying + tsk.Detail.Success + tsk.Detail.Queueing + tsk.Detail.Stopped
		taskRes.Data = append(taskRes.Data, tsk)
	}

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

func broadcastJobList() {
	for _, subscriber := range clientJobTaskSubscribers {
		meta, jobs := repo.findAllJob(subscriber.filter)

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
	res.TotalAlloc = fmt.Sprintf("%d MB", m.TotalAlloc/candihelper.MByte)
	res.NumGC = int(m.NumGC)
	res.NumGoroutines = runtime.NumGoroutine()
	return
}
