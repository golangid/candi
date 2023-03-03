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
	"sync"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
	"github.com/labstack/echo"
)

// tracerMiddleware for wrap from http inbound (request from client)
func (h *restServer) tracerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()

		isDisableTrace, _ := strconv.ParseBool(req.Header.Get(candihelper.HeaderDisableTrace))
		if isDisableTrace {
			c.SetRequest(req.WithContext(tracer.SkipTraceContext(req.Context())))
			return next(c)
		}

		operationName := fmt.Sprintf("%s %s", req.Method, req.Host)

		header := map[string]string{}
		for key := range c.Request().Header {
			header[key] = c.Request().Header.Get(key)
		}

		trace, ctx := tracer.StartTraceFromHeader(req.Context(), operationName, header)
		defer func() {
			trace.SetTag("trace_id", tracer.GetTraceID(ctx))
			trace.Finish()
			logger.LogGreen("rest_api > trace_url: " + tracer.GetTraceURL(ctx))
		}()

		httpDump, _ := httputil.DumpRequest(req, false)
		trace.SetTag("http.url_path", req.URL.Path)
		trace.SetTag("http.method", req.Method)
		trace.Log("http.request", httpDump)

		body, _ := io.ReadAll(req.Body)
		if len(body) < h.opt.jaegerMaxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			trace.Log("request.body", body)
		} else {
			trace.Log("request.body.size", len(body))
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body)) // reuse body

		resBody := new(bytes.Buffer)
		mw := io.MultiWriter(c.Response().Writer, resBody)
		c.Response().Writer = wrapper.NewWrapHTTPResponseWriter(mw, c.Response().Writer)
		c.SetRequest(req.WithContext(ctx))

		err := next(c)
		statusCode := c.Response().Status
		trace.SetTag("http.status_code", c.Response().Status)
		if statusCode >= http.StatusBadRequest {
			trace.SetError(fmt.Errorf("resp.code:%d", statusCode))
		}

		if resBody.Len() < h.opt.jaegerMaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
			trace.Log("response.body", resBody.String())
		} else {
			trace.Log("response.body.size", resBody.Len())
		}
		return err
	}
}

// EchoDefaultCORSMiddleware middleware
func EchoDefaultCORSMiddleware() echo.MiddlewareFunc {
	allowMethods := strings.Join(env.BaseEnv().CORSAllowMethods, ",")
	allowHeaders := strings.Join(env.BaseEnv().CORSAllowHeaders, ",")
	exposeHeaders := ""

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			req := c.Request()
			res := c.Response()
			origin := req.Header.Get(echo.HeaderOrigin)
			allowOrigin := ""

			// Check allowed origins
			for _, o := range env.BaseEnv().CORSAllowOrigins {
				if o == "*" && env.BaseEnv().CORSAllowCredential {
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
				res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
				res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowOrigin)
				if exposeHeaders != "" {
					res.Header().Set(echo.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				return next(c)
			}

			// Preflight request
			res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
			res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
			res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
			res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowOrigin)
			res.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods)
			if allowHeaders != "" {
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := req.Header.Get(echo.HeaderAccessControlRequestHeaders)
				if h != "" {
					res.Header().Set(echo.HeaderAccessControlAllowHeaders, h)
				}
			}
			return c.NoContent(http.StatusNoContent)
		}
	}
}

// EchoLoggerMiddleware middleware
func EchoLoggerMiddleware() echo.MiddlewareFunc {
	bPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {

			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			buf := bPool.Get().(*bytes.Buffer)
			buf.Reset()
			defer bPool.Put(buf)

			buf.WriteString(`{"time":"`)
			buf.WriteString(time.Now().Format(time.RFC3339Nano))

			buf.WriteString(`","id":"`)
			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			buf.WriteString(id)

			buf.WriteString(`","remote_ip":"`)
			buf.WriteString(c.RealIP())

			buf.WriteString(`","host":"`)
			buf.WriteString(req.Host)

			buf.WriteString(`","method":"`)
			buf.WriteString(req.Method)

			buf.WriteString(`","uri":"`)
			buf.WriteString(req.RequestURI)

			buf.WriteString(`","user_agent":"`)
			buf.WriteString(req.UserAgent())

			buf.WriteString(`","status":`)
			n := res.Status
			s := logger.GreenColor(n)
			switch {
			case n >= 500:
				s = logger.RedColor(n)
			case n >= 400:
				s = logger.YellowColor(n)
			case n >= 300:
				s = logger.CyanColor(n)
			}
			buf.WriteString(s)

			buf.WriteString(`,"error":"`)
			if err != nil {
				buf.WriteString(err.Error())
			}

			buf.WriteString(`","latency":`)
			l := stop.Sub(start)
			buf.WriteString(strconv.FormatInt(int64(l), 10))

			buf.WriteString(`,"latency_human":"`)
			buf.WriteString(stop.Sub(start).String())

			buf.WriteString(`","bytes_in":`)
			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}
			buf.WriteString(cl)

			buf.WriteString(`,"bytes_out":`)
			buf.WriteString(strconv.FormatInt(res.Size, 10))

			buf.WriteString("}\n")

			_, err = os.Stdout.Write(buf.Bytes())
			return
		}
	}
}
