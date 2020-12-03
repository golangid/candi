package taskqueueworker

import (
	"errors"
)

const maxClientSubscribers = 2

var errClientLimitExceeded = errors.New("client limit exceeded, please try again later")

func registerNewTaskListSubscriber(clientID string, clientChannel chan []TaskResolver) error {
	if len(clientTaskSubscribers) >= maxClientSubscribers {
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
	if len(clientJobTaskSubscribers) >= maxClientSubscribers {
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
	go broadcastTaskList()
	go broadcastJobList()
}

func broadcastTaskList() {
	var taskRes []TaskResolver
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
		taskRes = append(taskRes, tsk)
	}

	for _, subscriber := range clientTaskSubscribers {
		subscriber <- taskRes
	}
}

func broadcastJobList() {
	for _, subscriber := range clientJobTaskSubscribers {
		meta, jobs := repo.findAllJob(subscriber.filter)
		subscriber.c <- JobListResolver{
			Meta: meta,
			Data: jobs,
		}
	}
}
