package candiutils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gojektech/heimdall/v6"
	"github.com/gojektech/heimdall/v6/hystrix"
	"github.com/golangid/candi/tracer"
)

// httpRequestImpl struct
type httpRequestImpl struct {
	client *hystrix.Client

	breakerName               string
	timeout                   time.Duration
	retries                   int
	sleepBetweenRetry         time.Duration
	tlsConfig                 *tls.Config
	minHTTPErrorCodeThreshold int
}

// HTTPRequestOption func type
type HTTPRequestOption func(*httpRequestImpl)

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

// HTTPRequest interface
type HTTPRequest interface {
	Do(context context.Context, method, url string, reqBody []byte, headers map[string]string) ([]byte, int, error)
}

// NewHTTPRequest function
// Request's Constructor
// Returns : *Request
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

	// define a maximum jitter interval
	maximumJitterInterval := 10 * time.Millisecond
	// create a backoff
	backoff := heimdall.NewConstantBackoff(httpReq.sleepBetweenRetry, maximumJitterInterval)
	// create a new retry mechanism with the backoff
	retrier := heimdall.NewRetrier(backoff)

	hystrixClientOpt := []hystrix.Option{
		hystrix.WithHTTPTimeout(httpReq.timeout),
		hystrix.WithHystrixTimeout(httpReq.timeout),
		hystrix.WithRetrier(retrier),
		hystrix.WithRetryCount(httpReq.retries),
		hystrix.WithCommandName(httpReq.breakerName),
		hystrix.WithFallbackFunc(httpReq.fallbackErr),
	}
	if httpReq.tlsConfig != nil {
		hystrixClientOpt = append(hystrixClientOpt, hystrix.WithHTTPClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: httpReq.tlsConfig},
		}))
	}

	// set http client
	httpReq.client = hystrix.NewClient(hystrixClientOpt...)
	return httpReq
}

// Do function, for http client call
func (request *httpRequestImpl) Do(ctx context.Context, method, url string, requestBody []byte, headers map[string]string) (respBody []byte, respCode int, err error) {
	// set request http
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		tracer.SetError(ctx, err)
		return nil, 0, err
	}

	// set tracer
	trace, ctx := tracer.StartTraceWithContext(ctx, fmt.Sprintf("HTTP Request: %s %s", method, req.URL.Host))
	defer func() { trace.SetError(err); trace.Finish() }()

	// iterate optional data of headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	trace.InjectHTTPHeader(req)
	dumpRequest, _ := httputil.DumpRequest(req, false)
	trace.SetTag("http.request", dumpRequest)
	trace.SetTag("http.method", req.Method)
	trace.SetTag("http.url", req.URL.String())
	trace.SetTag("http.url_path", req.URL.Path)
	trace.SetTag("http.min_error_code", request.minHTTPErrorCodeThreshold)
	trace.SetTag("http.retries", request.retries)
	trace.SetTag("http.sleep_between_retry", request.sleepBetweenRetry.String())
	trace.SetTag("http.timeout", request.timeout.String())
	trace.SetTag("http.breaker_name", request.breakerName)
	if requestBody != nil {
		trace.Log("request.body", requestBody)
	}

	// client request
	resp, err := request.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	// close response body
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	respCode = resp.StatusCode

	dumpResponse, _ := httputil.DumpResponse(resp, false)
	trace.SetTag("response.header", dumpResponse)
	trace.SetTag("response.code", resp.StatusCode)
	trace.SetTag("response.status", resp.Status)
	trace.Log("response.body", respBody)

	if request.minHTTPErrorCodeThreshold != 0 && resp.StatusCode >= request.minHTTPErrorCodeThreshold {
		err = errors.New(resp.Status)
		var resp map[string]string
		json.Unmarshal(respBody, &resp)
		if resp["message"] != "" {
			err = errors.New(resp["message"])
		}
	}
	return
}

func (request *httpRequestImpl) fallbackErr(err error) error {
	// log error
	// ...
	return err
}
