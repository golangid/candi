package taskqueueworker

import (
	"context"
	"fmt"
	"strings"
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
	mongoJobCollections        = "task_queue_worker_jobs"
	mongoJobSummaryCollections = "task_queue_worker_job_summaries"
)

type mongoPersistent struct {
	db  *mongo.Database
	ctx context.Context
}

// NewMongoPersistent create mongodb persistent
func NewMongoPersistent(db *mongo.Database) Persistent {
	ctx := context.Background()

	uniqueOpts := &options.IndexOptions{
		Unique: candihelper.ToBoolPtr(true),
	}

	// check and create index in collection task_queue_worker_job_summaries
	indexViewJobSummaryColl := db.Collection(mongoJobSummaryCollections).Indexes()
	currentIndexSummaryNames := make(map[string]struct{})
	curJobSummary, err := indexViewJobSummaryColl.List(ctx)
	if err == nil {
		for curJobSummary.Next(ctx) {
			var result bson.M
			curJobSummary.Decode(&result)

			idxName, _ := result["name"].(string)
			if idxName != "" {
				currentIndexSummaryNames[idxName] = struct{}{}
			}
		}
	}

	indexes := map[string]mongo.IndexModel{
		"task_name_1": {
			Keys: bson.M{
				"task_name": 1,
			},
			Options: uniqueOpts,
		},
	}
	for name, idx := range indexes {
		if _, ok := currentIndexSummaryNames[name]; !ok {
			indexViewJobSummaryColl.CreateOne(ctx, idx)
		}
	}

	// check and create index in collection task_queue_worker_jobs
	indexViewJobColl := db.Collection(mongoJobCollections).Indexes()
	currentIndexNames := make(map[string]struct{})
	curJobColl, err := indexViewJobColl.List(ctx)
	if err == nil {
		for curJobColl.Next(ctx) {
			var result bson.M
			curJobColl.Decode(&result)

			idxName, _ := result["name"].(string)
			if idxName != "" {
				currentIndexNames[idxName] = struct{}{}
			}
		}
	}

	indexes = map[string]mongo.IndexModel{
		"task_name_1": {
			Keys: bson.M{
				"task_name": 1,
			},
			Options: &options.IndexOptions{},
		},
		"status_1": {
			Keys: bson.M{
				"status": 1,
			},
			Options: &options.IndexOptions{},
		},
		"created_at_1": {
			Keys: bson.M{
				"created_at": 1,
			},
			Options: &options.IndexOptions{},
		},
		"arguments_text_error_text": {
			Keys: bson.D{
				{Key: "arguments", Value: "text"},
				{Key: "error", Value: "text"},
			},
			Options: &options.IndexOptions{},
		},
		"task_name_1_status_1": {
			Keys: bson.D{
				{Key: "task_name", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: &options.IndexOptions{},
		},
		"task_name_1_created_at_1": {
			Keys: bson.D{
				{Key: "task_name", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: &options.IndexOptions{},
		},
		"task_name_1_status_1_created_at_1": {
			Keys: bson.D{
				{Key: "task_name", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: &options.IndexOptions{},
		},
	}

	for name, idx := range indexes {
		if _, ok := currentIndexNames[name]; !ok {
			indexViewJobColl.CreateOne(ctx, idx)
		}
	}

	return &mongoPersistent{
		db: db, ctx: ctx,
	}
}

func (s *mongoPersistent) FindAllJob(ctx context.Context, filter *Filter) (jobs []Job) {
	findOptions := &options.FindOptions{}

	if !filter.ShowAll {
		findOptions.SetLimit(int64(filter.Limit))
		findOptions.SetSkip(int64((filter.Page - 1) * filter.Limit))
	}
	if filter.Sort == "" {
		filter.Sort = "-created_at"
	}

	sort := 1
	if strings.HasPrefix(filter.Sort, "-") {
		sort = -1
	}
	findOptions.SetSort(bson.M{
		strings.TrimPrefix(filter.Sort, "-"): sort,
	})
	findOptions.SetProjection(bson.M{"retry_histories": 0})
	findOptions.SetAllowDiskUse(true)

	query := s.toBsonFilter(filter)
	cur, err := s.db.Collection(mongoJobCollections).Find(ctx, query, findOptions)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var job Job
		if err := cur.Decode(&job); err != nil {
			logger.LogE(err.Error())
			continue
		}
		if job.Status == string(statusSuccess) {
			job.Error = ""
		}
		if delay, err := time.ParseDuration(job.Interval); err == nil && job.Status == string(statusQueueing) {
			job.NextRetryAt = time.Now().Add(delay).In(candihelper.AsiaJakartaLocalTime).Format(time.RFC3339)
		}
		if job.TraceID != "" && defaultOption.tracingDashboard != "" {
			job.TraceID = fmt.Sprintf("%s/%s", defaultOption.tracingDashboard, job.TraceID)
		}
		job.CreatedAt = job.CreatedAt.In(candihelper.AsiaJakartaLocalTime)
		job.FinishedAt = job.FinishedAt.In(candihelper.AsiaJakartaLocalTime)
		if job.Retries > job.MaxRetry {
			job.Retries = job.MaxRetry
		}
		jobs = append(jobs, job)
	}

	return
}

func (s *mongoPersistent) CountAllJob(ctx context.Context, filter *Filter) int {
	queryFilter := s.toBsonFilter(filter)
	count, _ := s.db.Collection(mongoJobCollections).CountDocuments(ctx, queryFilter)
	return int(count)
}

func (s *mongoPersistent) AggregateAllTaskJob(ctx context.Context, filter *Filter) (results []TaskSummary) {

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
	findOptions.SetAllowDiskUse(true)
	csr, err := s.db.Collection(mongoJobCollections).Aggregate(ctx, pipeQuery, findOptions)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	defer csr.Close(ctx)

	csr.All(ctx, &results)
	return
}

func (s *mongoPersistent) FindAllSummary(ctx context.Context, filter *Filter) (result []TaskSummary) {

	query := bson.M{}

	if filter.TaskName != "" {
		query["task_name"] = filter.TaskName
	} else if len(filter.TaskNameList) > 0 {
		query["task_name"] = bson.M{
			"$in": filter.TaskNameList,
		}
	}

	cur, err := s.db.Collection(mongoJobSummaryCollections).Find(s.ctx, query)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &result); err != nil {
		logger.LogE(err.Error())
	}

	if len(filter.Statuses) > 0 {
		for i, res := range result {
			mapRes := res.ToMapResult()
			newCount := map[string]int{}
			for _, status := range filter.Statuses {
				newCount[strings.ToUpper(status)] = mapRes[strings.ToUpper(status)]
			}
			res.SetValue(newCount)
			result[i] = res
		}
	}

	return
}

func (s *mongoPersistent) IncrementSummary(ctx context.Context, taskName string, incr map[string]interface{}) {

	opt := options.UpdateOptions{
		Upsert: candihelper.ToBoolPtr(true),
	}
	for k, v := range incr {
		delete(incr, k)
		if k == "" {
			continue
		}
		incr[strings.ToLower(k)] = v
	}
	_, err := s.db.Collection(mongoJobSummaryCollections).UpdateOne(s.ctx,
		bson.M{
			"task_name": taskName,
		},
		bson.M{
			"$inc": incr,
		},
		&opt,
	)

	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *mongoPersistent) UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{}) {

	opt := options.UpdateOptions{
		Upsert: candihelper.ToBoolPtr(true),
	}
	for k, v := range updated {
		delete(updated, k)
		if k == "" {
			continue
		}
		updated[strings.ToLower(k)] = v
	}
	_, err := s.db.Collection(mongoJobSummaryCollections).UpdateOne(s.ctx,
		bson.M{
			"task_name": taskName,
		},
		bson.M{
			"$set": updated,
		},
		&opt,
	)

	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *mongoPersistent) SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) {
	tracer.Log(ctx, "persistent.mongo:save_job", job.ID)
	var err error

	if job.ID == "" {
		job.ID = primitive.NewObjectID().Hex()
		if len(job.RetryHistories) == 0 {
			job.RetryHistories = make([]RetryHistory, 0)
		}
		_, err = s.db.Collection(mongoJobCollections).InsertOne(ctx, job)
	} else {

		updateQuery := bson.M{
			"$set": bson.M(job.toMap()),
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
		_, err = s.db.Collection(mongoJobCollections).UpdateOne(ctx,
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

func (s *mongoPersistent) UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error) {

	updateQuery := bson.M{
		"$set": bson.M(updated),
	}
	if len(retryHistories) > 0 {
		updateQuery["$push"] = bson.M{
			"retry_histories": bson.M{
				"$each": retryHistories,
			},
		}
	}

	queryFilter := s.toBsonFilter(filter)
	res, err := s.db.Collection(mongoJobCollections).UpdateMany(ctx,
		queryFilter,
		updateQuery,
	)

	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}

	return res.MatchedCount, res.ModifiedCount, nil
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
	err = s.db.Collection(mongoJobCollections).FindOne(ctx, bson.M{"_id": id}, opts...).Decode(job)

	if len(job.RetryHistories) == 0 {
		job.RetryHistories = make([]RetryHistory, 0)
	}

	tracer.Log(ctx, "persistent.find_job_by_id", id)
	return
}

func (s *mongoPersistent) CleanJob(ctx context.Context, filter *Filter) (affectedRow int64) {

	res, err := s.db.Collection(mongoJobCollections).DeleteMany(ctx, s.toBsonFilter(filter))
	if err != nil {
		logger.LogE(err.Error())
		return affectedRow
	}

	return res.DeletedCount
}

func (s *mongoPersistent) DeleteJob(ctx context.Context, id string) (job Job, err error) {
	res := s.db.Collection(mongoJobCollections).FindOneAndDelete(ctx, bson.M{"_id": id})
	if res.Err() != nil {
		return job, res.Err()
	}
	res.Decode(&job)
	return
}

func (s *mongoPersistent) toBsonFilter(f *Filter) bson.M {
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

	if f.JobID != nil && *f.JobID != "" {
		pipeQuery = append(pipeQuery, bson.M{
			"_id": *f.JobID,
		})
	}
	if f.Search != nil && *f.Search != "" {
		pipeQuery = append(pipeQuery, bson.M{
			"$text": bson.M{"$search": fmt.Sprintf(`"%s"`, *f.Search)},
		})
	}
	if len(f.Statuses) > 0 {
		pipeQuery = append(pipeQuery, bson.M{
			"status": bson.M{
				"$in": f.Statuses,
			},
		})
	}
	if f.Status != nil {
		pipeQuery = append(pipeQuery, bson.M{
			"status": *f.Status,
		})
	}
	if !f.StartDate.IsZero() && !f.EndDate.IsZero() {
		pipeQuery = append(pipeQuery, bson.M{
			"created_at": bson.M{
				"$gte": f.StartDate, "$lte": f.EndDate,
			},
		})
	}

	if len(pipeQuery) > 0 {
		return bson.M{
			"$and": pipeQuery,
		}
	}

	return bson.M{}
}
