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
	// Deprecated: use InjectRequestHeader
	InjectHTTPHeader(req *http.Request)
	// Deprecated: use InjectRequestHeader
	InjectGRPCMetadata(md metadata.MD)
	InjectRequestHeader(header map[string]string)
	SetError(err error)
	Log(key string, value interface{})
	Finish(additionalTags ...map[string]interface{})
}
