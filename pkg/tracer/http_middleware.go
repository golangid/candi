package tracer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type httpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *httpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}
func (w *httpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// EchoRestTracerMiddleware for wrap from http inbound (request from client)
func EchoRestTracerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		tracer := opentracing.GlobalTracer()
		operationName := fmt.Sprintf("%s %s%s", req.Method, req.Host, req.URL.Path)

		var span opentracing.Span
		var ctx context.Context
		if spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
			span, ctx = opentracing.StartSpanFromContext(req.Context(), operationName)
			ext.SpanKindRPCServer.Set(span)
		} else {
			span = tracer.StartSpan(operationName, ext.RPCServerOption((spanCtx)))
			ctx = opentracing.ContextWithSpan(req.Context(), span)
			ext.SpanKindRPCClient.Set(span)
		}

		body, _ := ioutil.ReadAll(req.Body)
		if len(body) < maxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			span.SetTag("request.body", string(body))
		} else {
			span.SetTag("request.body.size", len(body))
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // reuse body

		span.SetTag("http.headers", req.Header)
		ext.HTTPUrl.Set(span, req.Host+req.RequestURI)
		ext.HTTPMethod.Set(span, req.Method)

		span.LogEvent("start_handling_request")

		defer func() {
			span.LogEvent("complete_handling_request")
			span.Finish()
			logger.LogGreen(GetTraceURL(ctx))
		}()

		resBody := new(bytes.Buffer)
		mw := io.MultiWriter(c.Response().Writer, resBody)
		writer := &httpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
		c.Response().Writer = writer
		c.SetRequest(req.WithContext(ctx))

		err := next(c)
		statusCode := c.Response().Status
		ext.HTTPStatusCode.Set(span, uint16(statusCode))
		if statusCode >= http.StatusBadRequest {
			ext.Error.Set(span, true)
		}

		if resBody.Len() < maxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
			span.SetTag("response.body", resBody.String())
		} else {
			span.SetTag("response.body.size", resBody.Len())
		}
		return err
	}
}
