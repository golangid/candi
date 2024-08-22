package restserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
)

// WithChainingMiddlewares chaining middlewares
func WithChainingMiddlewares(handlerFunc http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.HandlerFunc {
	if len(middlewares) == 0 {
		return handlerFunc
	}

	wrapped := handlerFunc
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](http.HandlerFunc(wrapped)).ServeHTTP
	}
	return wrapped
}

// HTTPMiddlewareCORS middleware for cors
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

			if allowCredential {
				res.Header().Set("Access-Control-Allow-Credentials", "true")
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

// HTTPMiddlewareTracer middleware wrapper for tracer
func HTTPMiddlewareTracer() func(http.Handler) http.Handler {
	bPool := candiutils.NewSyncPool(func() *bytes.Buffer {
		buff := bytes.NewBuffer(make([]byte, 256))
		buff.Reset()
		return buff
	}, func(b *bytes.Buffer) { b.Reset() })
	maxLogSize := env.BaseEnv().JaegerMaxPacketSize

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if _, ok := MiddlewareExcludeURLPath[req.URL.Path]; ok {
				next.ServeHTTP(rw, req)
				return
			}

			isDisableTrace, _ := strconv.ParseBool(req.Header.Get(candihelper.HeaderDisableTrace))
			if isDisableTrace {
				next.ServeHTTP(rw, req.WithContext(tracer.SkipTraceContext(req.Context())))
				return
			}

			header := make(map[string]string, len(req.Header))
			for key := range req.Header {
				header[key] = req.Header.Get(key)
			}

			trace, ctx := tracer.StartTraceFromHeader(req.Context(), "REST-Server", header)
			defer func() {
				if rec := recover(); rec != nil {
					trace.SetTag("panic", true)
					trace.SetError(fmt.Errorf("%v", rec))
					wrapper.NewHTTPResponse(http.StatusInternalServerError, "Something error").JSON(rw)
				}
				trace.Finish()
			}()

			httpDump, _ := httputil.DumpRequest(req, false)
			trace.Log("http.request", httpDump)
			trace.SetTag("http.host", req.Host)
			trace.SetTag("http.url_path", req.URL.Path)
			trace.SetTag("http.method", req.Method)

			if contentLength, err := strconv.Atoi(req.Header.Get("Content-Length")); err == nil {
				if contentLength < maxLogSize {
					reqBody := bPool.Get()
					reqBody.ReadFrom(req.Body)
					trace.Log("request.body", reqBody.String())
					req.Body = io.NopCloser(bytes.NewReader(reqBody.Bytes())) // reuse body
					bPool.Put(reqBody)
				} else {
					trace.Log("request.body.size", candihelper.TransformSizeToByte(uint64(contentLength)))
				}
			}

			start := time.Now()

			resBody := bPool.Get()
			defer bPool.Put(resBody)
			respWriter := wrapper.NewWrapHTTPResponseWriter(resBody, rw)
			respWriter.SetMaxWriteSize(maxLogSize)

			next.ServeHTTP(respWriter, req.WithContext(ctx))

			trace.SetTag("http.status_code", respWriter.StatusCode())
			if respWriter.StatusCode() >= http.StatusBadRequest {
				trace.SetError(fmt.Errorf("resp.code:%d", respWriter.StatusCode()))
			}
			trace.Log("response.header", respWriter.Header())
			if respWriter.GetContentLength() < maxLogSize {
				trace.Log("response.body", resBody.String())
			} else {
				trace.Log("response.body.size", candihelper.TransformSizeToByte(uint64(respWriter.GetContentLength())))
			}

			// log request
			stop := time.Now()
			logBuff := bPool.Get()
			defer bPool.Put(logBuff)

			logBuff.WriteString(`{"time":"`)
			logBuff.WriteString(time.Now().Format(time.RFC3339Nano))

			if id := req.Header.Get("X-Request-ID"); id != "" {
				logBuff.WriteString(`","id":"`)
				logBuff.WriteString(id)
			}

			logBuff.WriteString(`","remote_ip":"`)
			logBuff.WriteString(req.Header.Get("X-Real-IP"))

			logBuff.WriteString(`","method":"`)
			logBuff.WriteString(req.Method)

			logBuff.WriteString(`","host":"`)
			logBuff.WriteString(req.Host)

			logBuff.WriteString(`","uri":"`)
			logBuff.WriteString(req.RequestURI)

			logBuff.WriteString(`","user_agent":"`)
			logBuff.WriteString(req.UserAgent())

			logBuff.WriteString(`","status":`)

			s := logger.GreenColor(respWriter.StatusCode())
			switch {
			case respWriter.StatusCode() >= 500:
				s = logger.RedColor(respWriter.StatusCode())
			case respWriter.StatusCode() >= 400:
				s = logger.YellowColor(respWriter.StatusCode())
			case respWriter.StatusCode() >= 300:
				s = logger.CyanColor(respWriter.StatusCode())
			}
			logBuff.WriteString(s)

			logBuff.WriteString(`,"latency":"`)
			logBuff.WriteString(stop.Sub(start).String())

			logBuff.WriteString(`","bytes_in":`)
			cl := req.Header.Get("Content-Length")
			if cl == "" {
				cl = "0"
			}
			logBuff.WriteString(cl)

			logBuff.WriteString(`,"bytes_out":`)
			logBuff.WriteString(strconv.FormatInt(int64(respWriter.GetContentLength()), 10))

			logBuff.WriteString("}\n")

			io.Copy(os.Stdout, logBuff)
		})
	}
}
