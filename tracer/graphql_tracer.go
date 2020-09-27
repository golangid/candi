package tracer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	gqlerrors "github.com/golangid/graphql-go/errors"
	"github.com/golangid/graphql-go/introspection"
	"github.com/golangid/graphql-go/trace"
	"pkg.agungdwiprasetyo.com/gendon/helper"
	"pkg.agungdwiprasetyo.com/gendon/logger"
)

var gqlTypeNotShowLog = map[string]bool{
	"Query": true, "Mutation": true, "Subscription": true, "__Type": true, "__Schema": true,
}

// GraphQLTracer struct
type GraphQLTracer struct{}

// TraceQuery method
func (GraphQLTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {
	trace := StartTrace(ctx, fmt.Sprintf("GraphQL-Root:%s", operationName))

	tags := trace.Tags()
	tags["graphql.query"] = queryString
	tags["graphql.operationName"] = operationName
	if len(variables) != 0 {
		tags["graphql.variables"] = variables
	}

	return trace.Context(), func(errs []*gqlerrors.QueryError) {
		defer trace.Finish()
		logger.LogGreen(GetTraceURL(trace.Context()))

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
func (GraphQLTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	start := time.Now()
	return ctx, func(err *gqlerrors.QueryError) {
		end := time.Now()
		if !trivial && !gqlTypeNotShowLog[typeName] {
			statusColor := helper.Green
			status := " OK  "
			if err != nil {
				statusColor = helper.Red
				status = "ERROR"
			}

			arg, _ := json.Marshal(args)
			fmt.Fprintf(os.Stdout, "%s[GRAPHQL]%s => %s %10s %s | %v | %s %s %s | %13v | %s %s %s | %s\n",
				helper.White, helper.Reset,
				helper.Blue, typeName, helper.Reset,
				end.Format("2006/01/02 - 15:04:05"),
				statusColor, status, helper.Reset,
				end.Sub(start),
				helper.Magenta, label, helper.Reset,
				arg,
			)
		}
	}
}
