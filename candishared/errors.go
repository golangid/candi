package candishared

import (
	"context"
	"time"

	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/logger"
)

// GraphQLErrorResolver graphql error with extensions
type GraphQLErrorResolver interface {
	Error() string
	Extensions() map[string]interface{}
}

type resolveErrorImpl struct {
	message    string
	extensions map[string]interface{}
}

func (r *resolveErrorImpl) Error() string {
	return r.message
}
func (r *resolveErrorImpl) Extensions() map[string]interface{} {
	return r.extensions
}

// NewGraphQLErrorResolver constructor
func NewGraphQLErrorResolver(errMesage string, extensions map[string]interface{}) GraphQLErrorResolver {
	return &resolveErrorImpl{
		message: errMesage, extensions: extensions,
	}
}

// WorkerErrorHandler general function for handling error after execute worker handler
// example in this function can write log to database
func WorkerErrorHandler(ctx context.Context, workerType types.Worker, workerName string, message []byte, err error) {

	logger.LogYellow(string(workerType) + " - " + workerName + " - " + string(message) + " - handling error: " + string(err.Error()))
}

// TaskQueueErrorRetrier task queue worker for retry error
type TaskQueueErrorRetrier struct {
	Delay   time.Duration
	Message string
}

// Error implement error
func (e *TaskQueueErrorRetrier) Error() string {
	return e.Message
}
