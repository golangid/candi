package restserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
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
			span = globalTracer.StartSpan(operationName, ext.RPCServerOption((spanCtx)))
			ctx = opentracing.ContextWithSpan(req.Context(), span)
			ext.SpanKindRPCClient.Set(span)
		}

		body, _ := ioutil.ReadAll(req.Body)
		if len(body) < tracer.MaxPacketSize { // limit request body size to 65000 bytes (if higher tracer cannot show root span)
			span.LogKV("request.body", string(body))
		} else {
			span.SetTag("request.body.size", len(body))
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // reuse body

		span.SetTag("http.headers", string(candihelper.ToBytes(req.Header)))
		ext.HTTPUrl.Set(span, req.Host+req.RequestURI)
		ext.HTTPMethod.Set(span, req.Method)

		defer func() {
			span.Finish()
			logger.LogGreen("rest api " + tracer.GetTraceURL(ctx))
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

		if resBody.Len() < tracer.MaxPacketSize { // limit response body size to 65000 bytes (if higher tracer cannot show root span)
			span.LogKV("response.body", resBody.String())
		} else {
			span.SetTag("response.body.size", resBody.Len())
		}
		return err
	}
}
