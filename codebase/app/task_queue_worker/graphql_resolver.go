package taskqueueworker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	dashboard "github.com/golangid/candi-plugin/task-queue-worker/dashboard"
	"github.com/golangid/graphql-go"
	"github.com/golangid/graphql-go/relay"
	"github.com/golangid/graphql-go/trace"

	"github.com/golangid/candi"
	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/app/graphql_server/static"
	"github.com/golangid/candi/codebase/app/graphql_server/ws"
	"github.com/golangid/candi/config/env"
)

func serveGraphQLAPI(wrk *taskQueueWorker) {
	schemaOpts := []graphql.SchemaOpt{
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
		graphql.Tracer(trace.NoopTracer{}),
	}
	schema := graphql.MustParseSchema(schema, &rootResolver{worker: wrk}, schemaOpts...)

	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(dashboard.Dashboard)))
	mux.Handle("/task", http.StripPrefix("/", http.FileServer(dashboard.Dashboard)))
	mux.Handle("/job", http.StripPrefix("/", http.FileServer(dashboard.Dashboard)))

	mux.HandleFunc("/graphql", ws.NewHandlerFunc(schema, &relay.Handler{Schema: schema}))
	mux.HandleFunc("/voyager", func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte(static.VoyagerAsset)) })
	mux.HandleFunc("/playground", func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte(static.PlaygroundAsset)) })

	httpEngine := new(http.Server)
	httpEngine.Addr = fmt.Sprintf(":%d", defaultOption.dashboardPort)
	httpEngine.Handler = mux

	if err := httpEngine.ListenAndServe(); err != nil {
		switch e := err.(type) {
		case *net.OpError:
			panic(fmt.Errorf("task queue worker dashboard: %v", e))
		}
	}
}

type rootResolver struct {
	worker *taskQueueWorker
}

func (r *rootResolver) Tagline(ctx context.Context) (res TaglineResolver) {
	for taskClient := range clientTaskSubscribers {
		res.TaskListClientSubscribers = append(res.TaskListClientSubscribers, taskClient)
	}
	for client := range clientTaskJobListSubscribers {
		res.JobListClientSubscribers = append(res.JobListClientSubscribers, client)
	}
	res.Banner = defaultOption.dashboardBanner
	res.Tagline = "Task Queue Worker Dashboard"
	res.Version = candi.Version
	res.StartAt = env.BaseEnv().StartAt
	res.BuildNumber = env.BaseEnv().BuildNumber
	res.MemoryStatistics = getMemstats()
	return
}

func (r *rootResolver) GetJobDetail(ctx context.Context, input struct{ JobID string }) (res *Job, err error) {

	res, err = persistent.FindJobByID(ctx, input.JobID)
	if res.TraceID != "" && defaultOption.tracingDashboard != "" {
		res.TraceID = fmt.Sprintf("%s/%s", defaultOption.tracingDashboard, res.TraceID)
	}
	res.CreatedAt = res.CreatedAt.In(candihelper.AsiaJakartaLocalTime)
	if delay, err := time.ParseDuration(res.Interval); err == nil && res.Status == string(statusQueueing) {
		res.NextRetryAt = time.Now().Add(delay).In(candihelper.AsiaJakartaLocalTime).Format(time.RFC3339)
	}
	sort.Slice(res.RetryHistories, func(i, j int) bool {
		return res.RetryHistories[i].EndAt.After(res.RetryHistories[j].EndAt)
	})
	for i := range res.RetryHistories {
		res.RetryHistories[i].StartAt = res.RetryHistories[i].StartAt.In(candihelper.AsiaJakartaLocalTime)
		res.RetryHistories[i].EndAt = res.RetryHistories[i].EndAt.In(candihelper.AsiaJakartaLocalTime)

		if res.RetryHistories[i].TraceID != "" && defaultOption.tracingDashboard != "" {
			res.RetryHistories[i].TraceID = fmt.Sprintf("%s/%s", defaultOption.tracingDashboard, res.RetryHistories[i].TraceID)
		}
	}
	return
}

func (r *rootResolver) DeleteJob(ctx context.Context, input struct{ JobID string }) (ok string, err error) {
	err = persistent.DeleteJob(ctx, input.JobID)
	broadcastAllToSubscribers(r.worker.ctx)
	return
}

func (r *rootResolver) AddJob(ctx context.Context, input struct{ Param AddJobInputResolver }) (string, error) {

	job := AddJobRequest{
		TaskName: input.Param.TaskName,
		MaxRetry: int(input.Param.MaxRetry),
		Args:     []byte(input.Param.Args),
	}
	if input.Param.RetryInterval != nil {
		interval, err := time.ParseDuration(*input.Param.RetryInterval)
		if err != nil {
			return "", err
		}
		job.RetryInterval = interval
	}
	return AddJob(ctx, &job)
}

func (r *rootResolver) StopJob(ctx context.Context, input struct {
	JobID string
}) (string, error) {

	return "Success stop job " + input.JobID, StopJob(r.worker.ctx, input.JobID)
}

func (r *rootResolver) StopAllJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	if _, ok := registeredTask[input.TaskName]; !ok {
		return "", fmt.Errorf("task '%s' unregistered, task must one of [%s]", input.TaskName, strings.Join(tasks, ", "))
	}

	stopAllJobInTask(input.TaskName)
	queue.Clear(ctx, input.TaskName)
	persistent.UpdateJob(ctx,
		Filter{
			TaskName: input.TaskName, Status: []string{string(statusQueueing), string(statusRetrying)},
		},
		map[string]interface{}{
			"status": statusStopped,
		},
	)
	broadcastAllToSubscribers(r.worker.ctx)

	return "Success stop all job in task " + input.TaskName, nil
}

func (r *rootResolver) RetryJob(ctx context.Context, input struct {
	JobID string
}) (string, error) {

	return "Success retry job " + input.JobID, RetryJob(r.worker.ctx, input.JobID)
}

func (r *rootResolver) RetryAllJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	go func() {

		ctx := context.Background()

		filter := Filter{
			Page: 1, Limit: 10, Status: []string{string(statusFailure), string(statusStopped)}, TaskName: input.TaskName,
		}
		count := persistent.CountAllJob(ctx, filter)
		totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
		for filter.Page <= totalPages {
			jobs := persistent.FindAllJob(ctx, filter)
			for _, job := range jobs {
				job.Interval = defaultInterval.String()
				job.Retries = 0
				task := registeredTask[job.TaskName]
				job.Status = string(statusQueueing)
				queue.PushJob(ctx, &job)
				registerJobToWorker(&job, task.workerIndex)
			}
			filter.Page++
		}

		persistent.UpdateJob(ctx, filter, map[string]interface{}{
			"status":  statusQueueing,
			"retries": 0,
		})
		broadcastAllToSubscribers(r.worker.ctx)
		refreshWorkerNotif <- struct{}{}
	}()

	return "Success retry all failure job in task " + input.TaskName, nil
}

func (r *rootResolver) CleanJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	persistent.CleanJob(ctx, input.TaskName)
	broadcastAllToSubscribers(ctx)

	return "Success clean all job in task " + input.TaskName, nil
}

func (r *rootResolver) ClearAllClientSubscriber(ctx context.Context) (string, error) {

	for range clientTaskSubscribers {
		closeAllSubscribers <- struct{}{}
	}
	for range clientTaskJobListSubscribers {
		closeAllSubscribers <- struct{}{}
	}
	for range clientJobDetailSubscribers {
		closeAllSubscribers <- struct{}{}
	}

	return "Success clear all client subscriber", nil
}

func (r *rootResolver) GetAllActiveSubscriber(ctx context.Context) (cs []*ClientSubscriber, err error) {

	mapper := make(map[string]*ClientSubscriber)
	for k := range clientTaskSubscribers {
		_, ok := mapper[k]
		if !ok {
			mapper[k] = &ClientSubscriber{}
		}
		mapper[k].ClientID = k
		mapper[k].SubscribeList.TaskDashboard = true
	}
	for k, v := range clientTaskJobListSubscribers {
		_, ok := mapper[k]
		if !ok {
			mapper[k] = &ClientSubscriber{}
		}
		mapper[k].ClientID = k
		mapper[k].SubscribeList.JobList = v.filter
	}
	for k, v := range clientJobDetailSubscribers {
		_, ok := mapper[k]
		if !ok {
			mapper[k] = &ClientSubscriber{}
		}
		mapper[k].ClientID = k
		mapper[k].SubscribeList.JobDetailID = v.jobID
	}

	for _, v := range mapper {
		cs = append(cs, v)
	}

	sort.Slice(cs, func(i, j int) bool {
		return cs[i].ClientID < cs[j].ClientID
	})
	return cs, nil
}

func (r *rootResolver) ListenTaskDashboard(ctx context.Context) (<-chan TaskListResolver, error) {
	output := make(chan TaskListResolver)

	httpHeader := candishared.GetValueFromContext(ctx, candishared.ContextKeyHTTPHeader).(http.Header)
	clientID := httpHeader.Get("Sec-WebSocket-Key")

	if err := registerNewTaskListSubscriber(clientID, output); err != nil {
		return nil, err
	}

	autoRemoveClient := time.NewTicker(defaultOption.autoRemoveClientInterval)

	go broadcastTaskList(r.worker.ctx)

	go func() {
		defer func() { broadcastTaskList(r.worker.ctx); close(output); autoRemoveClient.Stop() }()

		select {
		case <-ctx.Done():
			removeTaskListSubscriber(clientID)
			return

		case <-closeAllSubscribers:
			output <- TaskListResolver{Meta: MetaTaskResolver{IsCloseSession: true}}
			removeTaskListSubscriber(clientID)
			return

		case <-autoRemoveClient.C:
			output <- TaskListResolver{
				Meta: MetaTaskResolver{
					IsCloseSession: true,
				},
			}
			removeTaskListSubscriber(clientID)
			return
		}

	}()

	return output, nil
}

func (r *rootResolver) ListenTaskJobList(ctx context.Context, input struct {
	TaskName           string
	Page, Limit        int32
	Search, JobID      *string
	Status             []string
	StartDate, EndDate *string
}) (<-chan JobListResolver, error) {

	output := make(chan JobListResolver)

	httpHeader := candishared.GetValueFromContext(ctx, candishared.ContextKeyHTTPHeader).(http.Header)
	clientID := httpHeader.Get("Sec-WebSocket-Key")

	if input.Page <= 0 {
		input.Page = 1
	}
	if input.Limit <= 0 || input.Limit > 10 {
		input.Limit = 10
	}

	filter := Filter{
		Page: int(input.Page), Limit: int(input.Limit), Search: input.Search, Status: input.Status, TaskName: input.TaskName,
		JobID: input.JobID,
	}

	filter.StartDate, _ = time.Parse(time.RFC3339, candihelper.PtrToString(input.StartDate))
	filter.EndDate, _ = time.Parse(time.RFC3339, candihelper.PtrToString(input.EndDate))

	if err := registerNewJobListSubscriber(input.TaskName, clientID, filter, output); err != nil {
		return nil, err
	}

	autoRemoveClient := time.NewTicker(defaultOption.autoRemoveClientInterval)

	go func() {
		jobs := persistent.FindAllJob(r.worker.ctx, filter)
		var meta MetaJobList
		filter.TaskNameList = []string{filter.TaskName}
		counterAll := persistent.AggregateAllTaskJob(r.worker.ctx, filter)
		if len(counterAll) == 1 {
			meta.Detail.Failure = counterAll[0].Detail.Failure
			meta.Detail.Retrying = counterAll[0].Detail.Retrying
			meta.Detail.Success = counterAll[0].Detail.Success
			meta.Detail.Queueing = counterAll[0].Detail.Queueing
			meta.Detail.Stopped = counterAll[0].Detail.Stopped
			meta.TotalRecords = counterAll[0].TotalJobs
		}
		meta.Page, meta.Limit = filter.Page, filter.Limit
		meta.TotalPages = int(math.Ceil(float64(meta.TotalRecords) / float64(meta.Limit)))
		candihelper.TryCatch{
			Try: func() {
				output <- JobListResolver{
					Meta: meta, Data: jobs,
				}
			},
			Catch: func(error) {},
		}.Do()
	}()

	go func() {
		defer func() { close(output); autoRemoveClient.Stop() }()

		select {
		case <-ctx.Done():
			removeJobListSubscriber(clientID)
			return

		case <-closeAllSubscribers:
			output <- JobListResolver{Meta: MetaJobList{IsCloseSession: true}}
			removeJobListSubscriber(clientID)
			return

		case <-autoRemoveClient.C:
			output <- JobListResolver{
				Meta: MetaJobList{
					IsCloseSession: true,
				},
			}
			removeJobListSubscriber(clientID)
			return

		}
	}()

	return output, nil
}

func (r *rootResolver) ListenJobDetail(ctx context.Context, input struct {
	JobID string
}) (<-chan Job, error) {

	output := make(chan Job)

	httpHeader := candishared.GetValueFromContext(ctx, candishared.ContextKeyHTTPHeader).(http.Header)
	clientID := httpHeader.Get("Sec-WebSocket-Key")

	if input.JobID == "" {
		return output, errors.New("Job ID cannot empty")
	}

	_, err := persistent.FindJobByID(ctx, input.JobID, "retry_histories")
	if err != nil {
		return output, errors.New("Job not found")
	}

	if err := registerNewJobDetailSubscriber(clientID, input.JobID, output); err != nil {
		return nil, err
	}

	autoRemoveClient := time.NewTicker(defaultOption.autoRemoveClientInterval)

	go func() {
		defer func() { close(output); autoRemoveClient.Stop() }()

		broadcastJobDetail(ctx)

		select {
		case <-ctx.Done():
			removeJobDetailSubscriber(clientID)
			return

		case <-closeAllSubscribers:
			removeJobDetailSubscriber(clientID)
			return

		case <-autoRemoveClient.C:
			removeJobDetailSubscriber(clientID)
			return

		}
	}()

	return output, nil
}
