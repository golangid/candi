package wrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

type HTTPMiddlewareTracerConfig struct {
	MaxLogSize  int
	ExcludePath map[string]struct{}
}

// HTTPMiddlewareTracer middleware wrapper for tracer
func HTTPMiddlewareTracer(cfg HTTPMiddlewareTracerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			if _, isExcludePath := cfg.ExcludePath[req.URL.Path]; isExcludePath {
				next.ServeHTTP(rw, req)
				return
			}
			isDisableTrace, _ := strconv.ParseBool(req.Header.Get(candihelper.HeaderDisableTrace))
			if isDisableTrace {
				next.ServeHTTP(rw, req.WithContext(tracer.SkipTraceContext(req.Context())))
				return
			}

			operationName := fmt.Sprintf("%s %s", req.Method, req.Host)

			header := map[string]string{}
			for key := range req.Header {
				header[key] = req.Header.Get(key)
			}

			trace, ctx := tracer.StartTraceFromHeader(req.Context(), operationName, header)
			defer func() {
				trace.SetTag("trace_id", tracer.GetTraceID(ctx))
				trace.Finish()
				logger.LogGreen("rest_server > trace_url: " + tracer.GetTraceURL(ctx))
			}()

			httpDump, _ := httputil.DumpRequest(req, false)
			trace.SetTag("http.url_path", req.URL.Path)
			trace.SetTag("http.method", req.Method)
			trace.Log("http.request", httpDump)

			body, _ := io.ReadAll(req.Body)
			if len(body) < cfg.MaxLogSize {
				trace.Log("request.body", body)
			} else {
				trace.Log("request.body.size", len(body))
			}
			req.Body = io.NopCloser(bytes.NewBuffer(body)) // reuse body

			resBody := &bytes.Buffer{}
			respWriter := NewWrapHTTPResponseWriter(resBody, rw)

			next.ServeHTTP(respWriter, req.WithContext(ctx))

			trace.SetTag("http.status_code", respWriter.statusCode)
			if respWriter.statusCode >= http.StatusBadRequest {
				trace.SetError(fmt.Errorf("resp.code:%d", respWriter.statusCode))
			}

			if resBody.Len() < cfg.MaxLogSize {
				trace.Log("response.body", resBody.String())
			} else {
				trace.Log("response.body.size", resBody.Len())
			}
		})
	}
}

// HTTPMiddlewareCORS middleware wrapper for tracer
func HTTPMiddlewareCORS(
	allowMethods, allowHeaders, allowOrigins []string,
	exposeHeaders []string,
	allowCredential bool,
) func(http.Handler) http.Handler {

	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}
	if len(allowMethods) == 0 {
		allowMethods = []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete}
	}
	exposeHeader := strings.Join(exposeHeaders, ",")

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

			origin := req.Header.Get("Origin")
			allowOrigin := ""

			// Check allowed origins
			for _, o := range allowOrigins {
				if o == "*" && allowCredential {
					allowOrigin = origin
					break
				}
				if o == "*" || o == origin {
					allowOrigin = o
					break
				}
			}

			// Simple request
			if req.Method != http.MethodOptions {
				res.Header().Add("Vary", "Origin")
				res.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				if exposeHeader != "" {
					res.Header().Set("Access-Control-Expose-Headers", exposeHeader)
				}
				next.ServeHTTP(res, req)
				return
			}

			// Preflight request
			res.Header().Add("Vary", "Origin")
			res.Header().Add("Vary", "Access-Control-Request-Method")
			res.Header().Add("Vary", "Access-Control-Request-Headers")
			res.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			res.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ","))
			if len(allowHeaders) > 0 {
				res.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ","))
			} else {
				h := req.Header.Get("Access-Control-Request-Headers")
				if h != "" {
					res.Header().Set("Access-Control-Allow-Headers", h)
				}
			}
			res.WriteHeader(http.StatusNoContent)
		})
	}
}

// HTTPHandlerDefaultRoot default root http handler
func HTTPHandlerDefaultRoot(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	payload := struct {
		BuildNumber string `json:"build_number,omitempty"`
		Message     string `json:"message,omitempty"`
		Hostname    string `json:"hostname,omitempty"`
		Timestamp   string `json:"timestamp,omitempty"`
		StartAt     string `json:"start_at,omitempty"`
		Uptime      string `json:"uptime,omitempty"`
	}{
		Message:   fmt.Sprintf("Service %s up and running", env.BaseEnv().ServiceName),
		Timestamp: now.Format(time.RFC3339Nano),
	}

	if startAt, err := time.Parse(time.RFC3339, env.BaseEnv().StartAt); err == nil {
		payload.StartAt = env.BaseEnv().StartAt
		payload.Uptime = now.Sub(startAt).String()
	}
	if env.BaseEnv().BuildNumber != "" {
		payload.BuildNumber = env.BaseEnv().BuildNumber
	}
	if hostname, err := os.Hostname(); err == nil {
		payload.Hostname = hostname
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

// HTTPHandlerMemstats calculate runtime statistic
func HTTPHandlerMemstats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	data := struct {
		NumGoroutine int         `json:"num_goroutine"`
		Memstats     interface{} `json:"memstats"`
	}{
		runtime.NumGoroutine(), m,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
