package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candiutils"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
)

// HTTPMiddlewareConfig config model
type HTTPMiddlewareConfig struct {
	MaxLogSize  int
	DisableFunc func(r *http.Request) bool
	Writer      io.Writer
	OnPanic     func(panicMessage interface{}) (respCode int, respMessage string)
}

// HTTPMiddlewareTracer middleware wrapper for tracer
func HTTPMiddlewareTracer(cfg HTTPMiddlewareConfig) func(http.Handler) http.Handler {
	bPool := candiutils.NewSyncPool(func() *bytes.Buffer {
		buff := bytes.NewBuffer(make([]byte, 256))
		buff.Reset()
		return buff
	}, func(b *bytes.Buffer) {
		b.Reset()
	})
	if cfg.OnPanic == nil {
		cfg.OnPanic = func(interface{}) (respCode int, respMessage string) {
			return http.StatusInternalServerError, "Something Error"
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if cfg.DisableFunc != nil && cfg.DisableFunc(req) {
				next.ServeHTTP(rw, req)
				return
			}

			isDisableTrace, _ := strconv.ParseBool(req.Header.Get(candihelper.HeaderDisableTrace))
			if isDisableTrace {
				next.ServeHTTP(rw, req.WithContext(tracer.SkipTraceContext(req.Context())))
				return
			}

			header := map[string]string{}
			for key := range req.Header {
				header[key] = req.Header.Get(key)
			}

			var err error
			trace, ctx := tracer.StartTraceFromHeader(req.Context(), "REST-Server", header)
			defer func() {
				if rec := recover(); rec != nil {
					trace.SetTag("panic", true)
					err = fmt.Errorf("%v", rec)
					wrapper.NewHTTPResponse(cfg.OnPanic(rec)).JSON(rw)
				}
				trace.SetTag("trace_id", tracer.GetTraceID(ctx))
				trace.Finish(tracer.FinishWithError(err))
				logger.LogGreen("rest_server > trace_url: " + tracer.GetTraceURL(ctx))
			}()

			httpDump, _ := httputil.DumpRequest(req, false)
			trace.Log("http.request", httpDump)
			trace.SetTag("http.host", req.Host)
			trace.SetTag("http.url_path", req.URL.Path)
			trace.SetTag("http.method", req.Method)

			if contentLength, err := strconv.Atoi(req.Header.Get("Content-Length")); err == nil {
				if contentLength < cfg.MaxLogSize {
					reqBody := bPool.Get()
					reqBody.ReadFrom(req.Body)
					trace.Log("request.body", reqBody.String())
					req.Body = io.NopCloser(bytes.NewReader(reqBody.Bytes())) // reuse body
					bPool.Put(reqBody)
				} else {
					trace.Log("request.body.size", candihelper.TransformSizeToByte(uint64(contentLength)))
				}
			}

			resBody := bPool.Get()
			defer bPool.Put(resBody)
			respWriter := wrapper.NewWrapHTTPResponseWriter(resBody, rw)
			respWriter.SetMaxWriteSize(cfg.MaxLogSize)

			next.ServeHTTP(respWriter, req.WithContext(ctx))

			trace.SetTag("http.status_code", respWriter.StatusCode())
			if respWriter.StatusCode() >= http.StatusBadRequest {
				err = fmt.Errorf("resp.code:%d", respWriter.StatusCode())
			}
			trace.Log("response.header", respWriter.Header())
			if respWriter.GetContentLength() < cfg.MaxLogSize {
				trace.Log("response.body", resBody.String())
			} else {
				trace.Log("response.body.size", candihelper.TransformSizeToByte(uint64(respWriter.GetContentLength())))
			}
		})
	}
}

// HTTPMiddlewareCORS middleware wrapper for cors
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

// HTTPMiddlewareLog middleware
func HTTPMiddlewareLog(cfg HTTPMiddlewareConfig) func(http.Handler) http.Handler {
	bPool := &sync.Pool{
		New: func() interface{} {
			buff := bytes.NewBuffer(make([]byte, 256))
			buff.Reset()
			return buff
		},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if cfg.DisableFunc != nil && cfg.DisableFunc(req) {
				next.ServeHTTP(res, req)
				return
			}

			start := time.Now()

			resBody := bytes.NewBuffer(make([]byte, 256))
			respWriter := wrapper.NewWrapHTTPResponseWriter(resBody, res)
			next.ServeHTTP(respWriter, req)

			stop := time.Now()
			buf := bPool.Get().(*bytes.Buffer)

			buf.WriteString(`{"time":"`)
			buf.WriteString(time.Now().Format(time.RFC3339Nano))

			buf.WriteString(`","id":"`)
			id := req.Header.Get("X-Request-ID")
			buf.WriteString(id)

			buf.WriteString(`","remote_ip":"`)
			buf.WriteString(req.Header.Get("X-Real-IP"))

			buf.WriteString(`","host":"`)
			buf.WriteString(req.Host)

			buf.WriteString(`","method":"`)
			buf.WriteString(req.Method)

			buf.WriteString(`","uri":"`)
			buf.WriteString(req.RequestURI)

			buf.WriteString(`","user_agent":"`)
			buf.WriteString(req.UserAgent())

			buf.WriteString(`","status":`)

			s := logger.GreenColor(respWriter.StatusCode())
			switch {
			case respWriter.StatusCode() >= 500:
				s = logger.RedColor(respWriter.StatusCode())
			case respWriter.StatusCode() >= 400:
				s = logger.YellowColor(respWriter.StatusCode())
			case respWriter.StatusCode() >= 300:
				s = logger.CyanColor(respWriter.StatusCode())
			}
			buf.WriteString(s)

			buf.WriteString(`","latency":`)
			l := stop.Sub(start)
			buf.WriteString(strconv.FormatInt(int64(l), 10))

			buf.WriteString(`,"latency_human":"`)
			buf.WriteString(stop.Sub(start).String())

			buf.WriteString(`","bytes_in":`)
			cl := req.Header.Get("Content-Length")
			if cl == "" {
				cl = "0"
			}
			buf.WriteString(cl)

			buf.WriteString(`,"bytes_out":`)
			buf.WriteString(strconv.FormatInt(int64(respWriter.GetContentLength()), 10))

			buf.WriteString("}\n")

			io.Copy(cfg.Writer, buf)
			bPool.Put(buf)
		})
	}
}
