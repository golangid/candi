package taskqueueworker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/agungdwiprasetyo/task-queue-worker-dashboard/external"
	"github.com/golangid/graphql-go"
	"github.com/golangid/graphql-go/relay"

	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/app/graphql_server/static"
	"pkg.agungdp.dev/candi/codebase/app/graphql_server/ws"
	"pkg.agungdp.dev/candi/config/env"
)

func serveGraphQLAPI(wrk *taskQueueWorker) {
	schemaOpts := []graphql.SchemaOpt{
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
	}
	schema := graphql.MustParseSchema(schema, &rootResolver{worker: wrk}, schemaOpts...)

	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(external.Dashboard)))
	mux.Handle("/task", http.StripPrefix("/task", http.FileServer(external.Dashboard)))
	mux.HandleFunc("/graphql", ws.NewHandlerFunc(schema, &relay.Handler{Schema: schema}))
	mux.HandleFunc("/voyager", func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte(static.VoyagerAsset)) })

	httpEngine := new(http.Server)
	httpEngine.Addr = fmt.Sprintf(":%d", env.BaseEnv().TaskQueueDashboardPort)
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
	res.Tagline = "Task Queue Worker Dashboard"
	res.MemoryStatistics = getMemstats()
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

	job, err := repo.findJobByID(input.JobID)
	if err != nil {
		return "Failed", err
	}

	job.Status = string(statusStopped)
	repo.saveJob(job)
	broadcastAllToSubscribers()

	return "Success stop job " + input.JobID, nil
}

func (r *rootResolver) StopAllJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	if _, ok := registeredTask[input.TaskName]; !ok {
		return "", fmt.Errorf("task '%s' unregistered, task must one of [%s]", input.TaskName, strings.Join(tasks, ", "))
	}

	queue.Clear(input.TaskName)
	repo.updateAllStatus(input.TaskName, statusStopped)
	broadcastAllToSubscribers()

	return "Success stop all job in task " + input.TaskName, nil
}

func (r *rootResolver) RetryJob(ctx context.Context, input struct {
	JobID string
}) (string, error) {

	job, err := repo.findJobByID(input.JobID)
	if err != nil {
		return "Failed", err
	}
	job.Interval = defaultInterval
	task := registeredTask[job.TaskName]
	go func(job Job) {
		if (job.Status == string(statusFailure)) || (job.Retries >= job.MaxRetry) {
			job.Retries = 0
		}
		job.Status = string(statusQueueing)
		queue.PushJob(&job)
		repo.saveJob(job)
		broadcastAllToSubscribers()
		registerJobToWorker(&job, task.workerIndex)
		refreshWorkerNotif <- struct{}{}
	}(job)

	return "Success retry job " + input.JobID, nil
}

func (r *rootResolver) CleanJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	repo.cleanJob(input.TaskName)
	go broadcastAllToSubscribers()

	return "Success clean all job in task " + input.TaskName, nil
}

func (r *rootResolver) ListenTask(ctx context.Context) (<-chan TaskListResolver, error) {
	output := make(chan TaskListResolver)

	httpHeader := candishared.GetValueFromContext(ctx, candishared.ContextKeyHTTPHeader).(http.Header)
	clientID := httpHeader.Get("Sec-WebSocket-Key")

	if err := registerNewTaskListSubscriber(clientID, output); err != nil {
		return nil, err
	}

	autoRemoveClient := time.NewTicker(defaultOption.AutoRemoveClientInterval)

	go func() {
		defer func() { close(output); autoRemoveClient.Stop() }()

		broadcastTaskList()

		select {
		case <-ctx.Done():
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

	autoRemoveClient := time.NewTicker(defaultOption.AutoRemoveClientInterval)

	go func() {
		defer func() { close(output); autoRemoveClient.Stop() }()

		meta, jobs := repo.findAllJob(filter)
		output <- JobListResolver{
			Meta: meta, Data: jobs,
		}

		select {
		case <-ctx.Done():
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
