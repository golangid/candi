package restserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/valyala/fasttemplate"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
	"pkg.agungdp.dev/candi/wrapper"
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

		globalTracer := opentracing.GlobalTracer()
		operationName := fmt.Sprintf("%s %s", req.Method, req.Host)

		var span opentracing.Span
		var ctx context.Context
		if spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
			span, ctx = opentracing.StartSpanFromContext(req.Context(), operationName)
			ext.SpanKindRPCServer.Set(span)
		} else {
			span = globalTracer.StartSpan(operationName, opentracing.ChildOf(spanCtx), ext.SpanKindRPCClient)
			ctx = opentracing.ContextWithSpan(req.Context(), span)
		}

		body, _ := ioutil.ReadAll(req.Body)
		if len(body) < h.opt.jaegerMaxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			span.LogKV("request.body", string(body))
		} else {
			span.LogKV("request.body.size", len(body))
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // reuse body

		httpDump, _ := httputil.DumpRequest(req, false)
		span.SetTag("http.request", string(httpDump))
		span.SetTag("http.url_path", req.URL.Path)
		ext.HTTPMethod.Set(span, req.Method)

		defer func() {
			span.Finish()
			logger.LogGreen("rest_api > trace_url: " + tracer.GetTraceURL(ctx))
		}()

		resBody := new(bytes.Buffer)
		mw := io.MultiWriter(c.Response().Writer, resBody)
		c.Response().Writer = wrapper.NewWrapHTTPResponseWriter(mw, c.Response().Writer)
		c.SetRequest(req.WithContext(ctx))

		err := next(c)
		statusCode := c.Response().Status
		ext.HTTPStatusCode.Set(span, uint16(statusCode))
		if statusCode >= http.StatusBadRequest {
			ext.Error.Set(span, true)
		}

		if resBody.Len() < h.opt.jaegerMaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
			span.LogKV("response.body", resBody.String())
		} else {
			span.LogKV("response.body.size", resBody.Len())
		}
		return err
	}
}

func defaultCORS() echo.MiddlewareFunc {
	allowMethods := strings.Join([]string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete}, ",")
	allowHeaders := ""
	exposeHeaders := ""

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			req := c.Request()
			res := c.Response()
			origin := req.Header.Get(echo.HeaderOrigin)
			allowOrigin := ""

			// Check allowed origins
			for _, o := range []string{"*"} {
				if o == "*" || o == origin {
					allowOrigin = o
					break
				}
			}

			res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
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
