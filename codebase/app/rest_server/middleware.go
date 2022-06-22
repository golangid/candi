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
	"github.com/labstack/gommon/color"
	"github.com/valyala/fasttemplate"
)

// echoRestTracerMiddleware for wrap from http inbound (request from client)
func (h *restServer) echoRestTracerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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

		body, _ := io.ReadAll(req.Body)
		if len(body) < h.opt.jaegerMaxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			trace.Log("request.body", body)
		} else {
			trace.Log("request.body.size", len(body))
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body)) // reuse body

		httpDump, _ := httputil.DumpRequest(req, false)
		trace.SetTag("http.request", string(httpDump))
		trace.SetTag("http.url_path", req.URL.Path)
		trace.SetTag("http.method", req.Method)

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

func defaultCORS() echo.MiddlewareFunc {
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

func echoLogger() echo.MiddlewareFunc {
	format := `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
		`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n"
	template := fasttemplate.New(format, "${", "}")
	colorer := color.New()
	colorer.SetOutput(os.Stdout)
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

			if _, err = template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_unix":
					return buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
				case "time_unix_nano":
					return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
				case "time_rfc3339":
					return buf.WriteString(time.Now().Format(time.RFC3339))
				case "time_rfc3339_nano":
					return buf.WriteString(time.Now().Format(time.RFC3339Nano))
				case "time_custom":
					return buf.WriteString(time.Now().Format(candihelper.TimeFormatLogger))
				case "id":
					id := req.Header.Get(echo.HeaderXRequestID)
					if id == "" {
						id = res.Header().Get(echo.HeaderXRequestID)
					}
					return buf.WriteString(id)
				case "remote_ip":
					return buf.WriteString(c.RealIP())
				case "host":
					return buf.WriteString(req.Host)
				case "uri":
					return buf.WriteString(req.RequestURI)
				case "method":
					return buf.WriteString(req.Method)
				case "path":
					p := req.URL.Path
					if p == "" {
						p = "/"
					}
					return buf.WriteString(p)
				case "protocol":
					return buf.WriteString(req.Proto)
				case "referer":
					return buf.WriteString(req.Referer())
				case "user_agent":
					return buf.WriteString(req.UserAgent())
				case "status":
					n := res.Status
					s := colorer.Green(n)
					switch {
					case n >= 500:
						s = colorer.Red(n)
					case n >= 400:
						s = colorer.Yellow(n)
					case n >= 300:
						s = colorer.Cyan(n)
					}
					return buf.WriteString(s)
				case "error":
					if err != nil {
						return buf.WriteString(err.Error())
					}
				case "latency":
					l := stop.Sub(start)
					return buf.WriteString(strconv.FormatInt(int64(l), 10))
				case "latency_human":
					return buf.WriteString(stop.Sub(start).String())
				case "bytes_in":
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						cl = "0"
					}
					return buf.WriteString(cl)
				case "bytes_out":
					return buf.WriteString(strconv.FormatInt(res.Size, 10))
				default:
					switch {
					case strings.HasPrefix(tag, "header:"):
						return buf.Write([]byte(c.Request().Header.Get(tag[7:])))
					case strings.HasPrefix(tag, "query:"):
						return buf.Write([]byte(c.QueryParam(tag[6:])))
					case strings.HasPrefix(tag, "form:"):
						return buf.Write([]byte(c.FormValue(tag[5:])))
					case strings.HasPrefix(tag, "cookie:"):
						cookie, err := c.Cookie(tag[7:])
						if err == nil {
							return buf.Write([]byte(cookie.Value))
						}
					}
				}
				return 0, nil
			}); err != nil {
				return
			}

			_, err = os.Stdout.Write(buf.Bytes())
			return
		}
	}
}
