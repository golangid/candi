package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/trace"
)

var (
	green        = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white        = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow       = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	red          = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue         = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta      = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan         = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

// PanicLogger struct
type PanicLogger struct {
}

// LogPanic panic logger
func (l *PanicLogger) LogPanic(_ context.Context, value interface{}) {
	fmt.Printf("%T\n", value)
}

// NoopTracer struct
type NoopTracer struct{}

// TraceQuery method
func (NoopTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {
	return ctx, func(errs []*errors.QueryError) {}
}

// TraceField method
func (NoopTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	start := time.Now()
	return ctx, func(err *errors.QueryError) {
		end := time.Now()
		if !trivial && typeName != "Query" {
			statusColor := green
			status := " OK  "
			if err != nil {
				statusColor = red
				status = "ERROR"
			}

			arg, _ := json.Marshal(args)
			fmt.Fprintf(os.Stdout, "%s[GRAPHQL]%s => %s %10s %s | %v | %s %s %s | %13v | %s %s %s | %s\n",
				white, reset,
				blue, typeName, reset,
				end.Format("2006/01/02 - 15:04:05"),
				statusColor, status, reset,
				end.Sub(start),
				magenta, label, reset,
				arg,
			)
		}
	}
}
