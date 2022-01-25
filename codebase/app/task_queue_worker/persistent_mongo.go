package taskqueueworker

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoColl = "task_queue_worker_jobs"
)

type mongoPersistent struct {
	db *mongo.Database
}

// NewMongoPersistent create mongodb persistent
func NewMongoPersistent(db *mongo.Database) Persistent {
	uniqueOpts := &options.IndexOptions{
		Unique: candihelper.ToBoolPtr(true),
	}
	indexes := []mongo.IndexModel{
		{
			Keys: bson.M{
				"_id": 1,
			},
			Options: uniqueOpts,
		},
		{
			Keys: bson.M{
				"task_name": 1,
			},
			Options: &options.IndexOptions{},
		},
		{
			Keys: bson.M{
				"status": 1,
			},
			Options: &options.IndexOptions{},
		},
		{
			Keys: bson.M{
				"arguments": "text",
			},
			Options: &options.IndexOptions{},
		},
		{
			Keys: bson.D{
				{Key: "task_name", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	}

	indexView := db.Collection(mongoColl).Indexes()
	for _, idx := range indexes {
		indexView.CreateOne(context.Background(), idx)
	}
	return &mongoPersistent{db}
}

func (s *mongoPersistent) FindAllJob(ctx context.Context, filter Filter) (jobs []Job) {
	findOptions := &options.FindOptions{
		Sort: bson.M{"created_at": -1},
	}

	if !filter.ShowAll {
		findOptions.SetLimit(int64(filter.Limit))
		findOptions.SetSkip(int64((filter.Page - 1) * filter.Limit))
	}
	findOptions.SetProjection(bson.M{"retry_histories": 0})

	cur, err := s.db.Collection(mongoColl).Find(ctx, s.toBsonFilter(filter), findOptions)
	if err != nil {
		return
	}
	for cur.Next(ctx) {
		var job Job
		cur.Decode(&job)
		if job.Status == string(statusSuccess) {
			job.Error = ""
		}
		if delay, err := time.ParseDuration(job.Interval); err == nil && job.Status == string(statusQueueing) {
			job.NextRetryAt = time.Now().Add(delay).In(candihelper.AsiaJakartaLocalTime).Format(time.RFC3339)
		}
		if job.TraceID != "" && defaultOption.jaegerTracingDashboard != "" {
			job.TraceID = fmt.Sprintf("%s/trace/%s", defaultOption.jaegerTracingDashboard, job.TraceID)
		}
		job.CreatedAt = job.CreatedAt.In(candihelper.AsiaJakartaLocalTime)
		job.FinishedAt = job.FinishedAt.In(candihelper.AsiaJakartaLocalTime)
		jobs = append(jobs, job)
	}

	return
}

func (s *mongoPersistent) CountAllJob(ctx context.Context, filter Filter) int {
	count, _ := s.db.Collection(mongoColl).CountDocuments(ctx, s.toBsonFilter(filter))
	return int(count)
}

func (s *mongoPersistent) AggregateAllTaskJob(ctx context.Context, filter Filter) (result []TaskResolver) {

	pipeQuery := []bson.M{
		{
			"$match": s.toBsonFilter(filter),
		},
		{
			"$project": bson.M{
				"task_name": "$task_name",
				"success":   bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", statusSuccess}}, "then": 1, "else": 0}},
				"queueing":  bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", statusQueueing}}, "then": 1, "else": 0}},
				"retrying":  bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", statusRetrying}}, "then": 1, "else": 0}},
				"failure":   bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", statusFailure}}, "then": 1, "else": 0}},
				"stopped":   bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", statusStopped}}, "then": 1, "else": 0}},
			},
		},
		{
			"$group": bson.M{
				"_id": "$task_name",
				"success": bson.M{
					"$sum": "$success",
				},
				"queueing": bson.M{
					"$sum": "$queueing",
				},
				"retrying": bson.M{
					"$sum": "$retrying",
				},
				"failure": bson.M{
					"$sum": "$failure",
				},
				"stopped": bson.M{
					"$sum": "$stopped",
				},
			},
		},
	}

	findOptions := options.Aggregate()
	findOptions.AllowDiskUse = candihelper.ToBoolPtr(true)
	csr, err := s.db.Collection(mongoColl).Aggregate(ctx, pipeQuery, findOptions)
	if err != nil {
		return
	}
	defer csr.Close(ctx)

	result = make([]TaskResolver, len(filter.TaskNameList))
	mapper := make(map[string]int, len(filter.TaskNameList))
	for i, task := range filter.TaskNameList {
		result[i].Name = task
		result[i].ModuleName = registeredTask[task].moduleName
		mapper[task] = i
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ModuleName < result[i].ModuleName
	})

	for csr.Next(ctx) {
		var obj struct {
			TaskName string `bson:"_id"`
			Success  int    `bson:"success"`
			Queueing int    `bson:"queueing"`
			Retrying int    `bson:"retrying"`
			Failure  int    `bson:"failure"`
			Stopped  int    `bson:"stopped"`
		}
		csr.Decode(&obj)

		if idx, ok := mapper[obj.TaskName]; ok {
			res := TaskResolver{
				Name:       obj.TaskName,
				ModuleName: registeredTask[obj.TaskName].moduleName,
				TotalJobs:  obj.Success + obj.Queueing + obj.Retrying + obj.Failure + obj.Stopped,
			}
			res.Detail.Success = obj.Success
			res.Detail.Queueing = obj.Queueing
			res.Detail.Retrying = obj.Retrying
			res.Detail.Failure = obj.Failure
			res.Detail.Stopped = obj.Stopped
			result[idx] = res
		}
	}

	return
}

func (s *mongoPersistent) SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) {
	tracer.Log(ctx, "persistent.mongo:save_job", job.ID)
	var err error

	if job.ID == "" {
		job.ID = primitive.NewObjectID().Hex()
		if len(job.RetryHistories) == 0 {
			job.RetryHistories = make([]RetryHistory, 0)
		}
		_, err = s.db.Collection(mongoColl).InsertOne(ctx, job)
	} else {

		updateQuery := bson.M{
			"$set": bson.M{
				"task_name":   job.TaskName,
				"arguments":   job.Arguments,
				"retries":     job.Retries,
				"max_retry":   job.MaxRetry,
				"interval":    job.Interval,
				"created_at":  job.CreatedAt,
				"finished_at": job.FinishedAt,
				"status":      job.Status,
				"error":       job.Error,
				"error_stack": job.ErrorStack,
				"trace_id":    job.TraceID,
			},
		}
		if len(retryHistories) > 0 {
			updateQuery["$push"] = bson.M{
				"retry_histories": bson.M{
					"$each": retryHistories,
				},
			}
		}

		opt := options.UpdateOptions{
			Upsert: candihelper.ToBoolPtr(true),
		}
		_, err = s.db.Collection(mongoColl).UpdateOne(ctx,
			bson.M{
				"_id": job.ID,
			},
			updateQuery,
			&opt)
	}

	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *mongoPersistent) UpdateAllStatus(ctx context.Context, taskName string, currentStatus []JobStatusEnum, updatedStatus JobStatusEnum) {
	filter := bson.M{}

	if taskName != "" {
		filter["task_name"] = taskName
	}
	filter["status"] = bson.M{"$in": currentStatus}
	_, err := s.db.Collection(mongoColl).UpdateMany(ctx,
		filter,
		bson.M{
			"$set": bson.M{"status": updatedStatus, "retries": 0},
		})

	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *mongoPersistent) FindJobByID(ctx context.Context, id string, excludeFields ...string) (job *Job, err error) {
	tracer.Log(ctx, "persistent.mongo:find_job", id)

	var opts []*options.FindOneOptions

	if len(excludeFields) > 0 {
		exclude := bson.M{}
		for _, exc := range excludeFields {
			exclude[exc] = 0
		}
		opts = append(opts, &options.FindOneOptions{
			Projection: exclude,
		})
	}

	job = &Job{}
	err = s.db.Collection(mongoColl).FindOne(ctx, bson.M{"_id": id}, opts...).Decode(job)

	if len(job.RetryHistories) == 0 {
		job.RetryHistories = make([]RetryHistory, 0)
	}

	tracer.Log(ctx, "persistent.find_job_by_id", id)
	return
}

func (s *mongoPersistent) CleanJob(ctx context.Context, taskName string) {

	query := bson.M{
		"$and": []bson.M{
			{"task_name": taskName},
			{"status": bson.M{"$nin": []JobStatusEnum{statusRetrying, statusQueueing}}},
		},
	}
	s.db.Collection(mongoColl).DeleteMany(ctx, query)
}

func (s *mongoPersistent) toBsonFilter(f Filter) bson.M {
	pipeQuery := []bson.M{}

	if f.TaskName != "" {
		pipeQuery = append(pipeQuery, bson.M{
			"task_name": f.TaskName,
		})
	} else if len(f.TaskNameList) > 0 {
		pipeQuery = append(pipeQuery, bson.M{
			"task_name": bson.M{
				"$in": f.TaskNameList,
			},
		})
	}

	if f.Search != nil && *f.Search != "" {
		pipeQuery = append(pipeQuery, bson.M{
			"arguments": primitive.Regex{Pattern: *f.Search, Options: "i"},
		})
	}
	if len(f.Status) > 0 {
		pipeQuery = append(pipeQuery, bson.M{
			"status": bson.M{
				"$in": f.Status,
			},
		})
	}

	return bson.M{
		"$and": pipeQuery,
	}
}

func (s *mongoPersistent) DeleteJob(ctx context.Context, id string) (err error) {
	_, err = s.db.Collection(mongoColl).DeleteOne(ctx, bson.M{"_id": id})
	return
}
