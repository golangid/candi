package taskqueueworker

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/agungdwiprasetyo/task-queue-worker-dashboard/external"
	"github.com/golangid/graphql-go"
	"github.com/golangid/graphql-go/relay"
	"github.com/gomodule/redigo/redis"

	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/candishared"
	"pkg.agungdwiprasetyo.com/candi/codebase/app/graphql_server/static"
	"pkg.agungdwiprasetyo.com/candi/codebase/app/graphql_server/ws"
	"pkg.agungdwiprasetyo.com/candi/config/env"
)

func newGraphQLAPI(wrk *taskQueueWorker) {
	schemaOpts := []graphql.SchemaOpt{
		graphql.UseStringDescriptions(),
		graphql.UseFieldResolvers(),
	}
	schema := graphql.MustParseSchema(schema, &rootResolver{worker: wrk, redisPool: wrk.service.GetDependency().GetRedisPool().WritePool()}, schemaOpts...)

	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(external.Dashboard)))
	mux.Handle("/task", http.StripPrefix("/task", http.FileServer(external.Dashboard)))
	mux.HandleFunc("/graphql", ws.NewHandlerFunc(schema, &relay.Handler{Schema: schema}))
	mux.HandleFunc("/playground", func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte(static.PlaygroundAsset)) })

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
	worker    *taskQueueWorker
	redisPool *redis.Pool
}

func (r *rootResolver) Tagline(ctx context.Context) string {
	return "GraphQL service for Task Queue Worker dashboard"
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

	conn := r.redisPool.Get()
	defer conn.Close()

	conn.Do("LREM", job.TaskName, 1, candihelper.ToBytes(job))

	job.Status = string(statusStopped)
	r.worker.broadcastEvent(&job)

	return "Success stop job " + input.JobID, nil
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
		if job.Status == string(statusFailure) {
			job.Retries = 0
		}
		job.Status = string(statusQueueing)
		repo.saveJob(job)
		queue.PushJob(&job)
		registerJobToWorker(&job, task.workerIndex)
		r.worker.broadcastRefreshClientSubscriber(&job)
	}(job)

	return "Success retry job " + input.JobID, nil
}

func (r *rootResolver) CleanJob(ctx context.Context, input struct {
	TaskName string
}) (string, error) {

	repo.cleanJob(input.TaskName)
	go r.worker.broadcastRefreshClientSubscriber(&Job{TaskName: input.TaskName})

	return "Success clean all job in task " + input.TaskName, nil
}

func (r *rootResolver) SubscribeAllTask(ctx context.Context) <-chan []TaskResolver {
	output := make(chan []TaskResolver)

	go func() {
		r.worker.listenUpdatedTask(output)
	}()

	return output
}

func (r *rootResolver) ListenTask(ctx context.Context, input struct {
	TaskName    string
	Page, Limit int32
	Search      *string
	Status      []string
}) <-chan JobListResolver {
	output := make(chan JobListResolver)

	go func() {

		httpHeader := candishared.GetValueFromContext(ctx, candishared.ContextKeyHTTPHeader).(http.Header)

		if input.Page <= 0 {
			input.Page = 1
		}
		if input.Limit <= 0 || input.Limit > 10 {
			input.Limit = 10
		}

		filter := Filter{
			Page: int(input.Page), Limit: int(input.Limit), Search: input.Search, Status: input.Status,
		}
		meta, jobs := repo.findAllJob(input.TaskName, filter)
		output <- JobListResolver{
			Meta: meta, Data: jobs,
		}

		r.worker.registerNewSubscriber(input.TaskName, filter, output)

		select {
		case <-ctx.Done():
			fmt.Println("close", httpHeader.Get("Sec-WebSocket-Key"))
			// close(output)
			return
		}
	}()

	return output
}
