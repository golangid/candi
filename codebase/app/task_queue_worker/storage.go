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
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/logger"
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
	}

	indexView := db.Collection(mongoColl).Indexes()
	for _, idx := range indexes {
		indexView.CreateOne(context.Background(), idx)
	}
}

type storage struct {
	mongoRead, mongoWrite *mongo.Database
}

func (s *storage) findAllJob(filter Filter) (meta Meta, jobs []Job) {
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
	cur, err := s.mongoRead.Collection(mongoColl).Find(ctx, query, findOptions)
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
			job.NextRetryAt = time.Now().Add(delay).Format(time.RFC3339)
		}
		if job.TraceID != "" && tracerHost != "" {
			job.TraceID = fmt.Sprintf("http://%s:16686/trace/%s", tracerHost, job.TraceID)
		}
		jobs = append(jobs, job)
	}

	meta.Detail.GiveUp = repo.countTaskJobDetail(filter.TaskName, statusFailure)
	meta.Detail.Retrying = repo.countTaskJobDetail(filter.TaskName, statusRetrying)
	meta.Detail.Success = repo.countTaskJobDetail(filter.TaskName, statusSuccess)
	meta.Detail.Queueing = repo.countTaskJobDetail(filter.TaskName, statusQueueing)
	meta.Detail.Stopped = repo.countTaskJobDetail(filter.TaskName, statusStopped)
	meta.TotalRecords = s.countTaskJob(filter)
	meta.Page, meta.Limit = filter.Page, filter.Limit
	meta.TotalPages = int(math.Ceil(float64(meta.TotalRecords) / float64(meta.Limit)))
	return
}

func (s *storage) countTaskJob(filter Filter) int {
	ctx := context.Background()

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

	count, _ := s.mongoRead.Collection(mongoColl).CountDocuments(ctx, query)
	return int(count)
}

func (s *storage) countTaskJobDetail(taskName string, status jobStatusEnum) int {
	ctx := context.Background()

	count, _ := s.mongoRead.Collection(mongoColl).CountDocuments(ctx, bson.M{"task_name": taskName, "status": status})
	return int(count)
}

func (s *storage) saveJob(job Job) {
	ctx := context.Background()
	var err error

	if job.ID == "" {
		job.ID = primitive.NewObjectID().Hex()
		_, err = s.mongoWrite.Collection(mongoColl).InsertOne(ctx, job)
	} else {
		opt := options.UpdateOptions{
			Upsert: candihelper.ToBoolPtr(true),
		}
		_, err = s.mongoWrite.Collection(mongoColl).UpdateOne(ctx,
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

func (s *storage) updateAllStatus(taskName string, status jobStatusEnum) {
	ctx := context.Background()
	filter := bson.M{
		"task_name": taskName,
	}
	if status == statusStopped {
		filter["status"] = bson.M{"$nin": []jobStatusEnum{statusFailure, statusSuccess, statusRetrying}}
	}
	_, err := s.mongoWrite.Collection(mongoColl).UpdateMany(ctx,
		filter,
		bson.M{
			"$set": bson.M{"status": string(status)},
		})

	if err != nil {
		logger.LogE(err.Error())
	}
}

func (s *storage) findJobByID(id string) (job Job, err error) {
	ctx := context.Background()

	err = s.mongoWrite.Collection(mongoColl).FindOne(ctx, bson.M{"_id": id}).Decode(&job)
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
	s.mongoWrite.Collection(mongoColl).DeleteMany(ctx, query)
}

func (s *storage) findAllPendingJob() (jobs []Job) {
	ctx := context.Background()

	query := bson.M{
		"status": bson.M{
			"$in": []jobStatusEnum{statusRetrying, statusQueueing},
		},
	}
	cur, err := s.mongoRead.Collection(mongoColl).Find(ctx, query)
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
