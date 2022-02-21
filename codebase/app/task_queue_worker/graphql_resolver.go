package taskqueueworker

import (
	"context"
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
	mux.HandleFunc("/task", func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, "/", http.StatusSeeOther)
	})
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
	for client := range clientJobTaskSubscribers {
		res.JobListClientSubscribers = append(res.JobListClientSubscribers, client)
	}
	res.Banner = defaultOption.dashboardBanner
	res.Tagline = "Task Queue Worker Dashboard"
	res.Version = candi.Version
	res.StartAt = startAt
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

func (r *rootResolver) AddJob(ctx context.Context, input struct {
	TaskName string
	MaxRetry int32
	Args     string
}) (string, error) {
	return "ok", AddJob(input.TaskName, int(input.MaxRetry), []byte(input.Args))
}

func (r *rootResolver) StopJob(ctx context.Context, input struct {
	JobID string
}) (string, error) {

	job, err := persistent.FindJobByID(ctx, input.JobID)
	if err != nil {
		return "Failed", err
	}

	job.Status = string(statusStopped)
	persistent.SaveJob(ctx, job)
	broadcastAllToSubscribers(r.worker.ctx)

	return "Success stop job " + input.JobID, nil
}

func (r *rootResolver) StopAllJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	if _, ok := registeredTask[input.TaskName]; !ok {
		return "", fmt.Errorf("task '%s' unregistered, task must one of [%s]", input.TaskName, strings.Join(tasks, ", "))
	}

	stopAllJobInTask(input.TaskName)
	queue.Clear(ctx, input.TaskName)
	persistent.UpdateAllStatus(ctx, input.TaskName, []JobStatusEnum{statusQueueing, statusRetrying}, statusStopped)
	broadcastAllToSubscribers(r.worker.ctx)

	return "Success stop all job in task " + input.TaskName, nil
}

func (r *rootResolver) RetryJob(ctx context.Context, input struct {
	JobID string
}) (string, error) {

	job, err := persistent.FindJobByID(ctx, input.JobID)
	if err != nil {
		return "Failed", err
	}
	job.Interval = defaultInterval
	task := registeredTask[job.TaskName]
	go func(job *Job) {
		if (job.Status == string(statusFailure)) || (job.Retries >= job.MaxRetry) {
			job.Retries = 0
		}
		job.Status = string(statusQueueing)
		queue.PushJob(ctx, job)
		persistent.SaveJob(r.worker.ctx, job)
		broadcastAllToSubscribers(r.worker.ctx)
		registerJobToWorker(job, task.workerIndex)
		refreshWorkerNotif <- struct{}{}
	}(job)

	return "Success retry job " + input.JobID, nil
}

func (r *rootResolver) CleanJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	persistent.CleanJob(ctx, input.TaskName)
	broadcastAllToSubscribers(ctx)

	return "Success clean all job in task " + input.TaskName, nil
}

func (r *rootResolver) RetryAllJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	pageNumber := 1
	filter := Filter{Limit: 10, Status: []string{string(statusFailure), string(statusStopped)}, TaskName: input.TaskName}
	count := persistent.CountAllJob(ctx, filter)
	totalPages := int(math.Ceil(float64(count) / float64(filter.Limit)))
	for pageNumber <= totalPages {
		filter.Page = pageNumber
		jobs := persistent.FindAllJob(ctx, filter)
		for _, job := range jobs {
			job.Interval = defaultInterval
			job.Retries = 0
			task := registeredTask[job.TaskName]
			job.Status = string(statusQueueing)
			queue.PushJob(ctx, &job)
			registerJobToWorker(&job, task.workerIndex)
		}
		pageNumber++
	}

	persistent.UpdateAllStatus(ctx, input.TaskName, []JobStatusEnum{statusFailure, statusStopped}, statusQueueing)
	go func() {
		broadcastAllToSubscribers(r.worker.ctx)
		refreshWorkerNotif <- struct{}{}
	}()

	return "Success retry all failure job in task " + input.TaskName, nil
}

func (r *rootResolver) ClearAllClientSubscriber(ctx context.Context) (string, error) {

	for range clientTaskSubscribers {
		closeAllSubscribers <- struct{}{}
	}
	for range clientJobTaskSubscribers {
		closeAllSubscribers <- struct{}{}
	}

	return "Success clear all client subscriber", nil
}

func (r *rootResolver) ListenTask(ctx context.Context) (<-chan TaskListResolver, error) {
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

func (r *rootResolver) ListenTaskJobDetail(ctx context.Context, input struct {
	TaskName    string
	Page, Limit int32
	Search      *string
	Status      []string
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
	}

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
			removeJobListSubscriber(input.TaskName, clientID)
			return

		case <-closeAllSubscribers:
			output <- JobListResolver{Meta: MetaJobList{IsCloseSession: true}}
			removeJobListSubscriber(input.TaskName, clientID)
			return

		case <-autoRemoveClient.C:
			output <- JobListResolver{
				Meta: MetaJobList{
					IsCloseSession: true,
				},
			}
			removeJobListSubscriber(input.TaskName, clientID)
			return

		}
	}()

	return output, nil
}
