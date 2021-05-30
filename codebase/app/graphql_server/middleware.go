package graphqlserver

/*
	GraphQL Middleware, intercept graphql request for middleware and tracing
*/

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	gqlerrors "github.com/golangid/graphql-go/errors"
	"github.com/golangid/graphql-go/introspection"
	"github.com/golangid/graphql-go/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

const schemaRootInstropectionField = "__schema"

var gqlTypeNotShowLog = map[string]bool{
	"Query": true, "Mutation": true, "Subscription": true, "__Type": true, "__Schema": true,
}

// graphqlMiddleware struct
type graphqlMiddleware struct {
	middleware types.MiddlewareGroup
}

// newGraphQLMiddleware constructor
func newGraphQLMiddleware(middleware types.MiddlewareGroup) *graphqlMiddleware {
	return &graphqlMiddleware{
		middleware: middleware,
	}
}

// TraceQuery method, intercept incoming query and add tracing
func (t *graphqlMiddleware) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {

	globalTracer := opentracing.GlobalTracer()
	traceName := strings.TrimSuffix(fmt.Sprintf("GraphQL-Root:%s", operationName), ":")

	headers, _ := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	var span opentracing.Span
	if spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(headers)); err != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, traceName)
		ext.SpanKindRPCServer.Set(span)
	} else {
		span = globalTracer.StartSpan(traceName, opentracing.ChildOf(spanCtx), ext.SpanKindRPCClient)
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	if len(headers) > 0 {
		span.SetTag("http.headers", string(candihelper.ToBytes(headers)))
	}
	span.SetTag("graphql.query", queryString)
	span.SetTag("graphql.operationName", operationName)
	if len(variables) > 0 {
		span.SetTag("graphql.variables", string(candihelper.ToBytes(variables)))
	}

	return ctx, func(data []byte, errs []*gqlerrors.QueryError) {
		defer span.Finish()

		if len(data) < tracer.MaxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			span.LogKV("data", string(data))
		} else {
			span.SetTag("data.size", len(data))
		}
		if len(errs) > 0 {
			span.LogKV("errors", string(candihelper.ToBytes(errs)))
			ext.Error.Set(span, true)
		}
		logger.LogGreen("graphql > trace_url: " + tracer.GetTraceURL(ctx))
	}
}

// TraceField method, intercept field per query and check middleware
func (t *graphqlMiddleware) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	start := time.Now()
	if middFunc, ok := t.middleware[fmt.Sprintf("%s.%s", typeName, fieldName)]; ok {
		for _, mw := range middFunc {
			ctx = mw(ctx)
		}
	}
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
