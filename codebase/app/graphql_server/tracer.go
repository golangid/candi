package graphqlserver

/*
	GraphQL Middleware, intercept graphql request for middleware and tracing
*/

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
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
}

// TraceQuery method, intercept incoming query and add tracing
func (t *graphqlTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, gqltrace.QueryFinishFunc) {
	headers, _ := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	header := map[string]string{}
	for key := range headers {
		header[key] = headers.Get(key)
	}

	trace, ctx := tracer.StartTraceFromHeader(ctx, "GraphQL-Root", header)
	if operationName != "" {
		trace.SetTag("graphql.operationName", operationName)
	}
	if len(headers) > 0 {
		trace.Log("http.headers", headers)
	}
	trace.Log("graphql.query", queryString)
	if len(variables) > 0 {
		trace.Log("graphql.variables", variables)
	}

	return ctx, func(data []byte, errs []*gqlerrors.QueryError) {
		if len(data) < env.BaseEnv().JaegerMaxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			trace.Log("response.data", data)
		} else {
			trace.Log("response.data.size", len(data))
		}
		if len(errs) > 0 {
			trace.Log("response.errors", errs)
			trace.SetError(errs[0])
		}
		logger.LogGreen("graphql > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish()
	}
}

// TraceField method, intercept field per query and check middleware
func (t *graphqlTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, gqltrace.FieldFinishFunc) {
	start := time.Now()
	return ctx, func(data []byte, err *gqlerrors.QueryError) {
		end := time.Now()
		if env.BaseEnv().DebugMode && !trivial && !gqlTypeNotShowLog[typeName] && fieldName != schemaRootInstropectionField {
			statusColor := candihelper.Green
			status := " OK  "
			if err != nil {
				statusColor = candihelper.Red
				status = "ERROR"
			}

			fmt.Fprintf(os.Stdout, "%s[GRAPHQL]%s => %s %10s %s | %v | %s %s %s | %13v | %s %s %s\n",
				candihelper.White, candihelper.Reset,
				candihelper.Blue, typeName, candihelper.Reset,
				end.Format("2006/01/02 - 15:04:05"),
				statusColor, status, candihelper.Reset,
				end.Sub(start),
				candihelper.Magenta, label, candihelper.Reset,
			)
		}
	}
}
