package taskqueueworker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoPersistent struct {
	db  *mongo.Database
	ctx context.Context

	summary Summary
}

// NewMongoPersistent create mongodb persistent
func NewMongoPersistent(db *mongo.Database) *MongoPersistent {
	ctx := context.Background()

	uniqueOpts := &options.IndexOptions{
		Unique: candihelper.ToBoolPtr(true),
	}

	// check and create index in collection task_queue_worker_job_summaries
	indexViewJobSummaryColl := db.Collection(jobSummaryModelName).Indexes()
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
	indexViewJobColl := db.Collection(jobModelName).Indexes()
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

	// check and create index in collection task_queue_worker_configurations
	indexViewConfigurationColl := db.Collection(configurationModelName).Indexes()
	currentIndexConfigurationNames := make(map[string]struct{})
	curConfiguration, err := indexViewConfigurationColl.List(ctx)
	if err == nil {
		for curConfiguration.Next(ctx) {
			var result bson.M
			curConfiguration.Decode(&result)

			idxName, _ := result["name"].(string)
			if idxName != "" {
				currentIndexConfigurationNames[idxName] = struct{}{}
			}
		}
	}
	indexes = map[string]mongo.IndexModel{
		"key_1": {
			Keys: bson.M{
				"key": 1,
			},
			Options: uniqueOpts,
		},
	}
	for name, idx := range indexes {
		if _, ok := currentIndexConfigurationNames[name]; !ok {
			indexViewConfigurationColl.CreateOne(ctx, idx)
		}
	}

	mp := &MongoPersistent{
		db: db, ctx: ctx,
	}

	// default summary from mongo
	mp.summary = mp
	return mp
}

func (s *MongoPersistent) SetSummary(summary Summary) {
	s.summary = summary
}
func (s *MongoPersistent) Summary() Summary {
	return s.summary
}

func (s *MongoPersistent) FindAllJob(ctx context.Context, filter *Filter) (jobs []Job) {
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

	query := s.toBsonFilter(filter)
	cur, err := s.db.Collection(jobModelName).Find(ctx, query, findOptions)
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
		jobs = append(jobs, job)
	}

	return
}

func (s *MongoPersistent) FindJobByID(ctx context.Context, id string, filterHistory *Filter) (job Job, err error) {
	var opts []*options.FindOneOptions
	filter := bson.M{"_id": id}

	if filterHistory == nil {
		opts = append(opts, &options.FindOneOptions{
			Projection: bson.M{"retry_histories": 0},
		})
	} else {
		opts = append(opts, &options.FindOneOptions{
			Projection: bson.M{"retry_histories": bson.M{"$slice": []interface{}{filterHistory.CalculateOffset(), filterHistory.Limit}}},
		})
		cur, err := s.db.Collection(jobModelName).Aggregate(ctx, []bson.M{
			{"$match": filter},
			{"$project": bson.M{"count": bson.M{"$size": "$retry_histories"}}},
		})
		if err == nil {
			for cur.Next(ctx) {
				var res struct {
					ID    string `bson:"_id"`
					Count int    `bson:"count"`
				}
				cur.Decode(&res)
				filterHistory.Count = res.Count
			}
			cur.Close(ctx)
		}
	}

	err = s.db.Collection(jobModelName).FindOne(ctx, filter, opts...).Decode(&job)
	if len(job.RetryHistories) == 0 {
		job.RetryHistories = make([]RetryHistory, 0)
	}

	return
}

func (s *MongoPersistent) CountAllJob(ctx context.Context, filter *Filter) int {
	queryFilter := s.toBsonFilter(filter)
	count, _ := s.db.Collection(jobModelName).CountDocuments(ctx, queryFilter)
	return int(count)
}

func (s *MongoPersistent) AggregateAllTaskJob(ctx context.Context, filter *Filter) (results []TaskSummary) {

	pipeQuery := []bson.M{
		{
			"$match": s.toBsonFilter(filter),
		},
		{
			"$project": bson.M{
				"task_name": "$task_name",
				"success":   bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", StatusSuccess}}, "then": 1, "else": 0}},
				"queueing":  bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", StatusQueueing}}, "then": 1, "else": 0}},
				"retrying":  bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", StatusRetrying}}, "then": 1, "else": 0}},
				"failure":   bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", StatusFailure}}, "then": 1, "else": 0}},
				"stopped":   bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", StatusStopped}}, "then": 1, "else": 0}},
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
	csr, err := s.db.Collection(jobModelName).Aggregate(ctx, pipeQuery, findOptions)
	if err != nil {
		logger.LogE(err.Error())
		return
	}
	defer csr.Close(ctx)

	csr.All(ctx, &results)
	return
}
func (s *MongoPersistent) SaveJob(ctx context.Context, job *Job, retryHistories ...RetryHistory) (err error) {
	tracer.Log(ctx, "persistent.mongo:save_job", job.ID)

	job.UpdatedAt = time.Now()
	if job.ID == "" {
		job.ID = uuid.New().String()
		job.CreatedAt = time.Now()
		if len(job.RetryHistories) == 0 {
			job.RetryHistories = make([]RetryHistory, 0)
		}
		_, err = s.db.Collection(jobModelName).InsertOne(ctx, job)
	} else {

		updated := job.toMap()
		updated["created_at"] = job.CreatedAt
		updateQuery := bson.M{
			"$set": bson.M(updated),
		}
		if len(retryHistories) > 0 {
			updateQuery["$push"] = bson.M{
				"retry_histories": bson.M{
					"$each": retryHistories,
					"$sort": bson.M{"start_at": -1},
				},
			}
		}

		opt := options.UpdateOptions{
			Upsert: candihelper.ToBoolPtr(true),
		}
		_, err = s.db.Collection(jobModelName).UpdateOne(ctx,
			bson.M{
				"_id": job.ID,
			},
			updateQuery,
			&opt)
	}

	if err != nil {
		logger.LogE(err.Error())
	}
	return
}

func (s *MongoPersistent) UpdateJob(ctx context.Context, filter *Filter, updated map[string]interface{}, retryHistories ...RetryHistory) (matchedCount, affectedRow int64, err error) {

	updated["updated_at"] = time.Now()
	updateQuery := bson.M{
		"$set": bson.M(updated),
	}
	if len(retryHistories) > 0 {
		updateQuery["$push"] = bson.M{
			"retry_histories": bson.M{
				"$each": retryHistories,
				"$sort": bson.M{"start_at": -1},
			},
		}
	}

	queryFilter := s.toBsonFilter(filter)
	res, err := s.db.Collection(jobModelName).UpdateMany(ctx,
		queryFilter,
		updateQuery,
	)

	if err != nil {
		logger.LogE(err.Error())
		return matchedCount, affectedRow, err
	}

	return res.MatchedCount, res.ModifiedCount, nil
}

func (s *MongoPersistent) CleanJob(ctx context.Context, filter *Filter) (affectedRow int64) {

	res, err := s.db.Collection(jobModelName).DeleteMany(ctx, s.toBsonFilter(filter))
	if err != nil {
		logger.LogE(err.Error())
		return affectedRow
	}

	return res.DeletedCount
}

func (s *MongoPersistent) DeleteJob(ctx context.Context, id string) (job Job, err error) {
	res := s.db.Collection(jobModelName).FindOneAndDelete(ctx, bson.M{"_id": id})
	if res.Err() != nil {
		return job, res.Err()
	}
	res.Decode(&job)
	return
}

func (s *MongoPersistent) toBsonFilter(f *Filter) bson.M {
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
	} else if len(f.ExcludeTaskNameList) > 0 {
		pipeQuery = append(pipeQuery, bson.M{
			"task_name": bson.M{
				"$nin": f.ExcludeTaskNameList,
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
	if startDate, endDate := f.ParseStartEndDate(); !startDate.IsZero() && !endDate.IsZero() {
		pipeQuery = append(pipeQuery, bson.M{
			"created_at": bson.M{
				"$gte": startDate, "$lte": endDate,
			},
		})
	}
	if f.BeforeCreatedAt != nil && !f.BeforeCreatedAt.IsZero() {
		pipeQuery = append(pipeQuery, bson.M{
			"created_at": bson.M{
				"$lte": *f.BeforeCreatedAt,
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

// summary

func (s *MongoPersistent) FindAllSummary(ctx context.Context, filter *Filter) (result []TaskSummary) {

	query := bson.M{}
	if filter.TaskName != "" {
		query["task_name"] = filter.TaskName
	} else if len(filter.TaskNameList) > 0 {
		query["task_name"] = bson.M{
			"$in": filter.TaskNameList,
		}
	}

	findOptions := &options.FindOptions{}
	findOptions.SetSort(bson.M{
		"task_name": 1,
	})

	cur, err := s.db.Collection(jobSummaryModelName).Find(s.ctx, query, findOptions)
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

func (s *MongoPersistent) FindDetailSummary(ctx context.Context, taskName string) (result TaskSummary) {
	s.db.Collection(jobSummaryModelName).FindOne(ctx, bson.M{"task_name": taskName}).Decode(&result)
	return
}

func (s *MongoPersistent) IncrementSummary(ctx context.Context, taskName string, incr map[string]int64) {
	if len(incr) == 0 {
		return
	}

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
	_, err := s.db.Collection(jobSummaryModelName).UpdateOne(s.ctx,
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

func (s *MongoPersistent) UpdateSummary(ctx context.Context, taskName string, updated map[string]interface{}) {

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
	_, err := s.db.Collection(jobSummaryModelName).UpdateOne(s.ctx,
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

func (s *MongoPersistent) DeleteAllSummary(ctx context.Context, filter *Filter) {
	_, err := s.db.Collection(jobSummaryModelName).DeleteMany(ctx, s.toBsonFilter(filter))
	if err != nil {
		logger.LogE(err.Error())
		return
	}
}

func (s *MongoPersistent) Ping(ctx context.Context) error {

	if err := s.db.Client().Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("mongodb ping: %v", err)
	}
	return nil
}

func (s *MongoPersistent) Type() string {
	var commandResult struct {
		Version string `bson:"version"`
	}
	err := s.db.RunCommand(s.ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&commandResult)
	logger.LogIfError(err)
	if commandResult.Version != "" {
		commandResult.Version = ", version: " + commandResult.Version
	}
	return "MongoDB Persistent" + commandResult.Version
}

func (s *MongoPersistent) GetAllConfiguration(ctx context.Context) (cfg []Configuration, err error) {
	cur, err := s.db.Collection(configurationModelName).Find(s.ctx, bson.M{})
	if err != nil {
		return cfg, err
	}
	defer cur.Close(ctx)
	cur.All(ctx, &cfg)
	return
}

func (s *MongoPersistent) GetConfiguration(key string) (cfg Configuration, err error) {
	err = s.db.Collection(configurationModelName).FindOne(s.ctx, bson.M{
		"key": key,
	}).Decode(&cfg)
	return
}

func (s *MongoPersistent) SetConfiguration(cfg *Configuration) (err error) {
	opt := options.UpdateOptions{
		Upsert: candihelper.ToBoolPtr(true),
	}
	_, err = s.db.Collection(configurationModelName).UpdateOne(s.ctx,
		bson.M{
			"key": cfg.Key,
		},
		bson.M{
			"$set": cfg,
		},
		&opt,
	)
	return
}
