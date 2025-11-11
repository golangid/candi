package graphqlserver

/*
	GraphQL Middleware, intercept graphql request for middleware and tracing
*/

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/tracer"
	gqlerrors "github.com/golangid/graphql-go/errors"
	"github.com/golangid/graphql-go/introspection"
	gqltrace "github.com/golangid/graphql-go/trace/tracer"
)

const schemaRootInstropectionField = "__schema"

var gqlTypeNotShowLog = map[string]bool{
	"Query": true, "Mutation": true, "Subscription": true, "__Type": true, "__Schema": true,
}

// graphqlTracer struct
type graphqlTracer struct {
	opt *Option
}

// TraceQuery method, intercept incoming query and add tracing
func (t *graphqlTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]any, varTypes map[string]*introspection.Type) (context.Context, gqltrace.QueryFinishFunc) {
	headers, _ := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	header := make(map[string]string, len(headers))
	for key := range headers {
		header[key] = headers.Get(key)
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "GraphQL", header)
	if operationName != "" {
		trace.SetTag("graphql.operationName", operationName)
	}
	if len(headers) > 0 {
		hw := bytes.NewBuffer(make([]byte, 0, 64))
		headers.Write(hw)
		trace.Log("http.headers", hw.String())
	}
	trace.Log("graphql.query", queryString)
	if len(variables) > 0 {
		trace.Log("graphql.variables", variables)
	}

	return ctx, func(data []byte, errs []*gqlerrors.QueryError) {
		if len(data) < int(t.opt.maxLogSize) {
			trace.Log("response.data", data)
		} else {
			trace.Log("response.data.size", len(data))
		}
		if len(errs) > 0 {
			trace.Log("response.errors", errs)
			trace.SetError(errs[0])
		}
		if t.opt.onErrorWrapper != nil {
			for i, err := range errs {
				if errWrap := t.opt.onErrorWrapper(ctx, err); errWrap != nil && errs[i] != nil {
					errs[i].Message = strings.TrimPrefix(errWrap.Error(), "graphql: ")
				}
			}
		}
		trace.Finish()
	}
}

// TraceField method, intercept field per query and check middleware
func (t *graphqlTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]any) (context.Context, gqltrace.FieldFinishFunc) {
	start := time.Now()
	return ctx, func(data []byte, err *gqlerrors.QueryError) {
		end := time.Now()
		if env.BaseEnv().DebugMode && !trivial && !gqlTypeNotShowLog[typeName] && fieldName != schemaRootInstropectionField {
			statusColor := []byte{27, 91, 57, 55, 59, 52, 50, 109} // green
			status := " OK  "
			if err != nil {
				statusColor = []byte{27, 91, 57, 55, 59, 52, 49, 109} // red
				status = "ERROR"
			}

			fmt.Fprintf(os.Stdout, "%s[GRAPHQL]%s => %s %10s %s | %v | %s %s %s | %5s | \x1b[35;1m%s\x1b[0m\n",
				[]byte{27, 91, 57, 48, 59, 52, 55, 109}, // white
				[]byte{27, 91, 48, 109},                 // reset
				[]byte{27, 91, 57, 55, 59, 52, 52, 109}, // blue
				typeName,
				[]byte{27, 91, 48, 109}, // reset
				end.Format("2006-01-02 15:04:05"),
				statusColor, status,
				[]byte{27, 91, 48, 109}, // reset
				end.Sub(start),
				label,
			)
		}
	}
}
