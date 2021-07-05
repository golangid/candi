package restserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
	"pkg.agungdp.dev/candi/wrapper"
)

// echoRestTracerMiddleware for wrap from http inbound (request from client)
func echoRestTracerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		globalTracer := opentracing.GlobalTracer()
		operationName := fmt.Sprintf("%s %s%s", req.Method, req.Host, req.URL.Path)

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
		if len(body) < env.BaseEnv().JaegerMaxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			span.LogKV("request.body", string(body))
		} else {
			span.LogKV("request.body.size", len(body))
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // reuse body

		httpDump, _ := httputil.DumpRequest(req, false)
		span.SetTag("http.request", string(httpDump))
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

		if resBody.Len() < env.BaseEnv().JaegerMaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
			span.LogKV("response.body", resBody.String())
		} else {
			span.LogKV("response.body.size", resBody.Len())
		}
		return err
	}
}
