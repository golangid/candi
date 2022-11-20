package taskqueueworker

import (
	"context"
	"runtime"
	"sort"
	"sync"

	"github.com/golangid/candi/candihelper"
)

type (
	subscriber struct {
		mutex                        sync.Mutex
		configuration                *configurationUsecase
		opt                          *option
		clientTaskSubscribers        map[string]*clientTaskDashboardSubscriber
		clientTaskJobListSubscribers map[string]*clientTaskJobListSubscriber
		clientJobDetailSubscribers   map[string]*clientJobDetailSubscriber
		closeAllSubscribers          chan struct{}
	}

	clientTaskDashboardSubscriber struct {
		c        chan TaskListResolver
		clientID string
		filter   *Filter
	}
	clientTaskJobListSubscriber struct {
		c             chan JobListResolver
		clientID      string
		skipBroadcast bool
		filter        *Filter
	}
	clientJobDetailSubscriber struct {
		c        chan JobResolver
		clientID string
		filter   *Filter
	}
)

func (s *clientTaskDashboardSubscriber) writeDataToChannel(data TaskListResolver) {
	defer func() { recover() }()
	s.c <- data
}
func (s *clientTaskJobListSubscriber) writeDataToChannel(data JobListResolver) {
	defer func() { recover() }()
	s.c <- data
}
func (s *clientJobDetailSubscriber) writeDataToChannel(data JobResolver) {
	defer func() { recover() }()
	s.c <- data
}

func initSubscriber(cfg *configurationUsecase, opt *option) *subscriber {
	return &subscriber{
		configuration:                cfg,
		opt:                          opt,
		clientTaskSubscribers:        make(map[string]*clientTaskDashboardSubscriber, cfg.getMaxClientSubscriber()),
		clientTaskJobListSubscribers: make(map[string]*clientTaskJobListSubscriber, cfg.getMaxClientSubscriber()),
		clientJobDetailSubscribers:   make(map[string]*clientJobDetailSubscriber, cfg.getMaxClientSubscriber()),
		closeAllSubscribers:          make(chan struct{}),
	}
}

func (s *subscriber) registerNewTaskListSubscriber(clientID string, filter *Filter, clientChannel chan TaskListResolver) error {
	if s.getTotalSubscriber() >= s.configuration.getMaxClientSubscriber() {
		return errClientLimitExceeded
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clientTaskSubscribers[clientID] = &clientTaskDashboardSubscriber{
		c: clientChannel, filter: filter, clientID: clientID,
	}
	return nil
}

func (s *subscriber) removeTaskListSubscriber(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.clientTaskSubscribers, clientID)
}

func (s *subscriber) registerNewJobListSubscriber(clientID string, filter *Filter, clientChannel chan JobListResolver) error {
	if s.getTotalSubscriber() >= s.configuration.getMaxClientSubscriber() {
		return errClientLimitExceeded
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clientTaskJobListSubscribers[clientID] = &clientTaskJobListSubscriber{
		c: clientChannel, filter: filter,
	}
	return nil
}

func (s *subscriber) removeJobListSubscriber(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.clientTaskJobListSubscribers, clientID)
}

func (s *subscriber) registerNewJobDetailSubscriber(clientID string, filter *Filter, clientChannel chan JobResolver) error {
	if s.getTotalSubscriber() >= s.configuration.getMaxClientSubscriber() {
		return errClientLimitExceeded
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clientJobDetailSubscribers[clientID] = &clientJobDetailSubscriber{
		c: clientChannel, filter: filter,
	}
	return nil
}

func (s *subscriber) removeJobDetailSubscriber(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.clientJobDetailSubscribers, clientID)
}

func (s *subscriber) getTotalSubscriber() int {
	return len(s.clientTaskSubscribers) + len(s.clientTaskJobListSubscribers) + len(s.clientJobDetailSubscribers)
}

func (s *subscriber) broadcastAllToSubscribers(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	go func(ctx context.Context) {
		if len(s.clientTaskSubscribers) > 0 {
			s.broadcastTaskList(ctx)
		}
		if len(s.clientTaskJobListSubscribers) > 0 {
			s.broadcastJobList(ctx)
		}
		if len(s.clientJobDetailSubscribers) > 0 {
			s.broadcastJobDetail(ctx)
		}
	}(ctx)
}

func (s *subscriber) broadcastTaskList(ctx context.Context) {

	var taskRes TaskListResolver
	taskRes.Data = make([]TaskResolver, 0)
	for _, summary := range s.opt.persistent.Summary().FindAllSummary(ctx, &Filter{}) {
		taskRes.Data = append(taskRes.Data, summary.ToTaskResolver())
	}

	sort.Slice(taskRes.Data, func(i, j int) bool {
		return taskRes.Data[i].ModuleName < taskRes.Data[i].ModuleName
	})

	taskRes.Meta.TotalRecords = len(taskRes.Data)
	taskRes.Meta.TotalClientSubscriber = s.getTotalSubscriber()

	for clientID, subscriber := range s.clientTaskSubscribers {
		taskRes.Meta.Page = subscriber.filter.Page
		taskRes.Meta.Limit = subscriber.filter.Limit
		taskRes.Meta.CalculatePage()
		taskRes.Meta.ClientID = clientID
		subscriber.writeDataToChannel(taskRes)
	}
}

func (s *subscriber) broadcastJobList(ctx context.Context) {
	for clientID := range s.clientTaskJobListSubscribers {
		s.broadcastJobListToClient(ctx, clientID)
	}
}

func (s *subscriber) broadcastJobListToClient(ctx context.Context, clientID string) {

	subscriber, ok := s.clientTaskJobListSubscribers[clientID]
	if !ok {
		return
	}

	if subscriber.filter.TaskName != "" {
		summary := s.opt.persistent.Summary().FindDetailSummary(ctx, subscriber.filter.TaskName)
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
		(subscriber.filter.StartDate != "" && subscriber.filter.EndDate != "")

	var jobListResolver JobListResolver
	jobListResolver.GetAllJob(ctx, subscriber.filter)
	jobListResolver.Meta.IsFreezeBroadcast = subscriber.skipBroadcast
	subscriber.writeDataToChannel(jobListResolver)
}

func (s *subscriber) broadcastJobDetail(ctx context.Context) {

	for clientID, subscriber := range s.clientJobDetailSubscribers {
		detail, err := s.opt.persistent.FindJobByID(ctx, candihelper.PtrToString(subscriber.filter.JobID), subscriber.filter)
		if err != nil {
			s.removeJobDetailSubscriber(clientID)
			continue
		}
		var jobResolver JobResolver
		jobResolver.ParseFromJob(&detail, 200)
		jobResolver.Meta.Page = subscriber.filter.Page
		jobResolver.Meta.TotalHistory = subscriber.filter.Count
		subscriber.writeDataToChannel(jobResolver)
	}
}

func (s *subscriber) broadcastWhenChangeAllJob(ctx context.Context, taskName string, isLoading bool, loadingMessage string) {

	s.opt.persistent.Summary().UpdateSummary(ctx, taskName, map[string]interface{}{
		"is_loading": isLoading, "loading_message": loadingMessage,
	})

	var taskRes TaskListResolver
	taskRes.Data = make([]TaskResolver, 0)
	for _, summary := range s.opt.persistent.Summary().FindAllSummary(ctx, &Filter{}) {
		taskRes.Data = append(taskRes.Data, summary.ToTaskResolver())
	}

	sort.Slice(taskRes.Data, func(i, j int) bool {
		return taskRes.Data[i].ModuleName < taskRes.Data[i].ModuleName
	})

	taskRes.Meta.TotalClientSubscriber = s.getTotalSubscriber()

	for _, subscriber := range s.clientTaskSubscribers {
		subscriber.writeDataToChannel(taskRes)
	}
	for _, subscriber := range s.clientTaskJobListSubscribers {
		subscriber.skipBroadcast = isLoading
		subscriber.writeDataToChannel(JobListResolver{
			Meta: MetaJobList{IsLoading: isLoading},
		})
	}
}

func getMemstats() (res MemstatsResolver) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	res.Alloc = m.Alloc
	res.TotalAlloc = m.TotalAlloc
	res.NumGC = int(m.NumGC)
	res.NumGoroutines = runtime.NumGoroutine()
	return
}
