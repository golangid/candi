package graphqlserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	gqlerrors "github.com/golangid/graphql-go/errors"
	"github.com/golangid/graphql-go/introspection"
	"github.com/golangid/graphql-go/trace"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/candishared"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

const schemaRootInstropectionField = "__schema"

var gqlTypeNotShowLog = map[string]bool{
	"Query": true, "Mutation": true, "Subscription": true, "__Type": true, "__Schema": true,
}

// graphQLTracer struct
type graphQLTracer struct {
	midd types.GraphQLMiddlewareGroup
}

// newGraphQLTracer constructor
func newGraphQLTracer(midd types.GraphQLMiddlewareGroup) *graphQLTracer {
	return &graphQLTracer{
		midd: midd,
	}
}

// TraceQuery method
func (t *graphQLTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {
	trace := tracer.StartTrace(ctx, strings.TrimSuffix(fmt.Sprintf("GraphQL-Root:%s", operationName), ":"))

	tags := trace.Tags()
	tags["graphql.query"] = queryString
	tags["graphql.operationName"] = operationName
	if len(variables) != 0 {
		tags["graphql.variables"] = variables
	}

	if headers, ok := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header); ok {
		tags["http.header"] = headers
	}

	return trace.Context(), func(data []byte, errs []*gqlerrors.QueryError) {
		defer trace.Finish()
		logger.LogGreen("graphql " + tracer.GetTraceURL(trace.Context()))
		tags["response.data"] = string(data)

		if len(errs) > 0 {
			tags["errors"] = errs
			msg := errs[0].Error()
			if len(errs) > 1 {
				msg += fmt.Sprintf(" (and %d more errors)", len(errs)-1)
			}
			trace.SetError(errors.New(msg))
		}
	}
}

// TraceField method
func (t *graphQLTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	start := time.Now()
	if middFunc, ok := t.midd[fmt.Sprintf("%s.%s", typeName, fieldName)]; ok {
		ctx = middFunc(ctx)
	}
	return ctx, func(data []byte, err *gqlerrors.QueryError) {
		end := time.Now()
		if !trivial && !gqlTypeNotShowLog[typeName] && fieldName != schemaRootInstropectionField {
			statusColor := candihelper.Green
			status := " OK  "
			if err != nil {
				statusColor = candihelper.Red
				status = "ERROR"
			}

			arg, _ := json.Marshal(args)
			fmt.Fprintf(os.Stdout, "%s[GRAPHQL]%s => %s %10s %s | %v | %s %s %s | %13v | %s %s %s | %s\n",
				candihelper.White, candihelper.Reset,
				candihelper.Blue, typeName, candihelper.Reset,
				end.Format("2006/01/02 - 15:04:05"),
				statusColor, status, candihelper.Reset,
				end.Sub(start),
				candihelper.Magenta, label, candihelper.Reset,
				arg,
			)
		}
	}
}
