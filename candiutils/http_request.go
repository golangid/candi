package candiutils

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/golangid/candi/tracer"
)

type (
	// HTTPRequest interface
	HTTPRequest interface {
		DoRequest(ctx context.Context, method, url string, requestBody []byte, headers map[string]string) (result *HTTPRequestResult, err error)
		Do(context context.Context, method, url string, reqBody []byte, headers map[string]string) (respBody []byte, respCode int, err error)
	}

	httpClientDo interface {
		Do(req *http.Request) (*http.Response, error)
	}

	// httpRequestImpl struct
	httpRequestImpl struct {
		client httpClientDo

		breakerName               string
		timeout                   time.Duration
		retries                   int
		sleepBetweenRetry         time.Duration
		tlsConfig                 *tls.Config
		minHTTPErrorCodeThreshold int
	}

	// HTTPRequestResult struct
	HTTPRequestResult struct {
		*bytes.Buffer
		RespCode int
	}

	// HTTPRequestOption func type
	HTTPRequestOption func(*httpRequestImpl)
)

// HTTPRequestSetRetries option func
func HTTPRequestSetRetries(retries int) HTTPRequestOption {
	return func(h *httpRequestImpl) {
		h.retries = retries
	}
}

// HTTPRequestSetSleepBetweenRetry option func
func HTTPRequestSetSleepBetweenRetry(sleepBetweenRetry time.Duration) HTTPRequestOption {
	return func(h *httpRequestImpl) {
		h.sleepBetweenRetry = sleepBetweenRetry
	}
}

// HTTPRequestSetTLS option func
func HTTPRequestSetTLS(tlsConfig *tls.Config) HTTPRequestOption {
	return func(h *httpRequestImpl) {
		h.tlsConfig = tlsConfig
	}
}

// HTTPRequestSetHTTPErrorCodeThreshold option func, set minimum http response code for return error when exec client request
func HTTPRequestSetHTTPErrorCodeThreshold(minHTTPStatusCode int) HTTPRequestOption {
	return func(h *httpRequestImpl) {
		h.minHTTPErrorCodeThreshold = minHTTPStatusCode
	}
}

// HTTPRequestSetTimeout option func
func HTTPRequestSetTimeout(timeout time.Duration) HTTPRequestOption {
	return func(h *httpRequestImpl) {
		h.timeout = timeout
	}
}

// HTTPRequestSetBreakerName option func
func HTTPRequestSetBreakerName(breakerName string) HTTPRequestOption {
	return func(h *httpRequestImpl) {
		h.breakerName = breakerName
	}
}

// HTTPRequestSetClient option func
func HTTPRequestSetClient(cl *http.Client) HTTPRequestOption {
	return func(h *httpRequestImpl) {
		h.client = cl
	}
}

// NewHTTPRequest function
// Request's Constructor
func NewHTTPRequest(opts ...HTTPRequestOption) HTTPRequest {
	httpReq := new(httpRequestImpl)

	// set default value
	httpReq.retries = 5
	httpReq.sleepBetweenRetry = 500 * time.Millisecond
	httpReq.minHTTPErrorCodeThreshold = http.StatusBadRequest
	httpReq.timeout = 10 * time.Second
	httpReq.breakerName = "default"

	for _, o := range opts {
		o(httpReq)
	}

	// set http client
	if httpReq.client == nil {
		client := &http.Client{
			Timeout: httpReq.timeout,
		}
		if httpReq.tlsConfig != nil {
			client.Transport = &http.Transport{TLSClientConfig: httpReq.tlsConfig}
		}
		httpReq.client = client
	}

	return httpReq
}

// Do function, for http client call
func (req *httpRequestImpl) Do(ctx context.Context, method, url string, requestBody []byte, headers map[string]string) (respBody []byte, respCode int, err error) {
	httpResult, err := req.DoRequest(ctx, method, url, requestBody, headers)
	if err != nil {
		if httpResult != nil {
			respBody, respCode = httpResult.Bytes(), httpResult.RespCode
		}
		return respBody, respCode, err
	}

	return httpResult.Bytes(), httpResult.RespCode, nil
}

func (req *httpRequestImpl) DoRequest(ctx context.Context, method, url string, requestBody []byte, headers map[string]string) (result *HTTPRequestResult, err error) {
	// set request http
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		tracer.SetError(ctx, err)
		return nil, err
	}

	// set tracer
	trace, ctx := tracer.StartTraceWithContext(ctx, fmt.Sprintf("HTTP Request: %s %s", method, httpReq.URL.Host))
	defer func() { trace.Finish(tracer.FinishWithError(err)) }()

	if headers == nil {
		headers = map[string]string{}
	}
	trace.InjectRequestHeader(headers)

	// iterate optional data of headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	trace.SetTag("http.method", httpReq.Method)
	trace.SetTag("http.url", httpReq.URL.String())
	trace.SetTag("http.url_path", httpReq.URL.Path)
	trace.SetTag("http.timeout", req.timeout.String())
	trace.SetTag("http.breaker_name", req.breakerName)

	dumpRequest, _ := httputil.DumpRequest(httpReq, false)
	trace.SetTag("http.request", dumpRequest)
	if requestBody != nil {
		trace.Log("request.body", requestBody)
	}

	resp, err := req.client.Do(httpReq)
	if err != nil && resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	result = &HTTPRequestResult{
		Buffer:   &bytes.Buffer{},
		RespCode: resp.StatusCode,
	}
	io.Copy(result.Buffer, resp.Body)

	dumpResponse, _ := httputil.DumpResponse(resp, false)
	trace.SetTag("http.response", dumpResponse)
	trace.SetTag("response.code", resp.StatusCode)
	trace.SetTag("response.status", resp.Status)
	trace.Log("response.body", result.Bytes())

	if req.minHTTPErrorCodeThreshold != 0 && resp.StatusCode >= req.minHTTPErrorCodeThreshold {
		err = errors.New(resp.Status)
	}
	return
}
