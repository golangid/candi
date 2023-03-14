package wrapper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

// HTTPMiddlewareTracer middleware wrapper for tracer
func HTTPMiddlewareTracer(maxLogSize int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			isDisableTrace, _ := strconv.ParseBool(req.Header.Get(candihelper.HeaderDisableTrace))
			if isDisableTrace || req.URL.Path == "/" {
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
			if len(body) < maxLogSize {
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

			if resBody.Len() < maxLogSize {
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
