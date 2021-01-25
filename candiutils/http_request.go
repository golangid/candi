package candiutils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/httpclient"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

// httpRequestImpl struct
type httpRequestImpl struct {
	client *httpclient.Client

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
	httpReq.minHTTPErrorCodeThreshold = http.StatusInternalServerError

	for _, o := range opts {
		o(httpReq)
	}

	// define a maximum jitter interval
	maximumJitterInterval := 5 * time.Millisecond

	// create a backoff
	backoff := heimdall.NewConstantBackoff(httpReq.sleepBetweenRetry, maximumJitterInterval)

	// create a new retry mechanism with the backoff
	retrier := heimdall.NewRetrier(backoff)

	// set http timeout
	timeout := 10000 * time.Millisecond

	heimdallClientOpt := []httpclient.Option{
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(httpReq.retries),
	}
	if httpReq.tlsConfig != nil {
		heimdallClientOpt = append(heimdallClientOpt, httpclient.WithHTTPClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: httpReq.tlsConfig},
		}))
	}

	// set http client
	client := httpclient.NewClient(heimdallClientOpt...)

	return &httpRequestImpl{
		client: client,
	}
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
	trace := tracer.StartTrace(ctx, fmt.Sprintf("HTTP Request: %s %s%s", method, req.URL.Host, req.URL.Path))
	defer func() {
		if err != nil {
			trace.SetError(err)
		}
		trace.Finish()
	}()
	ctx = trace.Context()

	// iterate optional data of headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	trace.InjectHTTPHeader(req)
	trace.SetTag("http.headers", req.Header)
	trace.SetTag("http.method", req.Method)
	trace.SetTag("http.url", req.URL.String())
	if requestBody != nil {
		tracer.Log(ctx, "request.body", requestBody)
	}

	// client request
	resp, err := request.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	// close response body
	defer resp.Body.Close()

	respBody, err = ioutil.ReadAll(resp.Body)
	respCode = resp.StatusCode

	trace.SetTag("response.code", resp.StatusCode)
	trace.SetTag("response.status", resp.Status)
	trace.SetTag("response.header", resp.Header)
	tracer.Log(ctx, "response.body", respBody)

	if request.minHTTPErrorCodeThreshold != 0 && resp.StatusCode >= request.minHTTPErrorCodeThreshold {
		err = errors.New(resp.Status)
		var resp map[string]string
		json.Unmarshal(respBody, &resp)
		if resp["message"] != "" {
			err = errors.New(resp["message"])
		}
		return
	}

	return
}
