package interfaces

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// Tracer for trace
type Tracer interface {
	Context() context.Context
	Tags() map[string]interface{}
	SetTag(key string, value interface{})
	InjectHTTPHeader(req *http.Request)
	InjectGRPCMetadata(md metadata.MD)
	SetError(err error)
	Log(key string, value interface{})
	Finish(additionalTags ...map[string]interface{})
}
