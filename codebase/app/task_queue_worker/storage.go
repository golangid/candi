package taskqueueworker

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/logger"
)

const (
	mongoColl = "task_queue_worker_jobs"
)

func createMongoIndex(db *mongo.Database) {
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
}

type storage struct {
	db *mongo.Database
}

func (s *storage) findAllJob(filter Filter) (meta MetaJobList, jobs []Job) {
	ctx := context.Background()

	lim := int64(filter.Limit)
	offset := int64((filter.Page - 1) * filter.Limit)
	findOptions := &options.FindOptions{
		Limit: &lim,
		Skip:  &offset,
		Sort:  bson.M{"created_at": -1},
	}

	pipeQuery := []bson.M{
		{"task_name": filter.TaskName},
	}
	if filter.Search != nil && *filter.Search != "" {
		pipeQuery = append(pipeQuery, bson.M{
			"arguments": primitive.Regex{Pattern: *filter.Search, Options: "i"},
		})
	}
	if len(filter.Status) > 0 {
		pipeQuery = append(pipeQuery, bson.M{
			"status": bson.M{
				"$in": filter.Status,
			},
		})
	}
	query := bson.M{
		"$and": pipeQuery,
	}
	cur, err := s.db.Collection(mongoColl).Find(ctx, query, findOptions)
	if err != nil {
		fmt.Println(err)
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
		if job.TraceID != "" && defaultOption.JaegerTracingDashboard != "" {
			job.TraceID = fmt.Sprintf("%s/trace/%s", defaultOption.JaegerTracingDashboard, job.TraceID)
		}
		job.CreatedAt = job.CreatedAt.In(candihelper.AsiaJakartaLocalTime)
		job.FinishedAt = job.FinishedAt.In(candihelper.AsiaJakartaLocalTime)
		jobs = append(jobs, job)
	}

	filter.TaskNameList = []string{filter.TaskName}
	counterAll := repo.countAllJobTask(context.Background(), filter)
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
	return
}

func (s *storage) findAllFailureJob(filter Filter) (jobs []Job) {
	ctx := context.Background()
	lim := int64(filter.Limit)
	offset := int64((filter.Page - 1) * filter.Limit)
	findOptions := &options.FindOptions{
		Limit: &lim,
		Skip:  &offset,
		Sort:  bson.M{"created_at": -1},
	}

	pipeQuery := []bson.M{
		{"task_name": filter.TaskName},
	}
	if len(filter.Status) > 0 {
		pipeQuery = append(pipeQuery, bson.M{
			"status": bson.M{
				"$in": filter.Status,
			},
		})
	}
	query := bson.M{
		"$and": pipeQuery,
	}
	cur, err := s.db.Collection(mongoColl).Find(ctx, query, findOptions)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	cur.All(ctx, &jobs)
	return
}

func (s *storage) countTaskJob(filter Filter) int {
	ctx := context.Background()

	count, _ := s.db.Collection(mongoColl).CountDocuments(ctx, filter.toBsonFilter())
	return int(count)
}

func (s *storage) countAllJobTask(ctx context.Context, filter Filter) (result []TaskResolver) {

	pipeQuery := []bson.M{
		{
			"$match": filter.toBsonFilter(),
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
		mapper[task] = i
	}

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
				Name:      obj.TaskName,
				TotalJobs: obj.Success + obj.Queueing + obj.Retrying + obj.Failure + obj.Stopped,
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

func (s *storage) saveJob(job *Job) {
	ctx := context.Background()
	var err error

	if job.ID == "" {
		job.ID = primitive.NewObjectID().Hex()
		_, err = s.db.Collection(mongoColl).InsertOne(ctx, job)
	} else {
		opt := options.UpdateOptions{
			Upsert: candihelper.ToBoolPtr(true),
		}
		_, err = s.db.Collection(mongoColl).UpdateOne(ctx,
			bson.M{
				"_id": job.ID,
			},
			bson.M{
				"$set": job,
			}, &opt)
	}

	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *storage) updateAllStatus(taskName string, status jobStatusEnum, currentStatus []jobStatusEnum) {
	ctx := context.Background()
	filter := bson.M{
		"task_name": taskName,
	}
	filter["status"] = bson.M{"$in": currentStatus}
	_, err := s.db.Collection(mongoColl).UpdateMany(ctx,
		filter,
		bson.M{
			"$set": bson.M{"status": string(status), "retries": 0},
		})

	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *storage) pauseAllRunningJob() {
	ctx := context.Background()
	_, err := s.db.Collection(mongoColl).UpdateMany(ctx,
		bson.M{
			"status": statusRetrying,
		},
		bson.M{
			"$set": bson.M{"status": string(statusQueueing)},
		})
	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *storage) findJobByID(id string) (job *Job, err error) {
	ctx := context.Background()

	job = &Job{}
	err = s.db.Collection(mongoColl).FindOne(ctx, bson.M{"_id": id}).Decode(job)
	return
}

func (s *storage) cleanJob(taskName string) {
	ctx := context.Background()

	query := bson.M{
		"$and": []bson.M{
			{"task_name": taskName},
			{"status": bson.M{"$nin": []jobStatusEnum{statusRetrying, statusQueueing}}},
		},
	}
	s.db.Collection(mongoColl).DeleteMany(ctx, query)
}

func (s *storage) findAllPendingJob() (jobs []Job) {
	ctx := context.Background()

	query := bson.M{
		"status": bson.M{
			"$in": []jobStatusEnum{statusRetrying, statusQueueing},
		},
	}
	cur, err := s.db.Collection(mongoColl).Find(ctx, query)
	if err != nil {
		return
	}
	for cur.Next(ctx) {
		var job Job
		cur.Decode(&job)
		jobs = append(jobs, job)
	}
	return
}
