package candiutils

import (
	"bytes"
	"context"
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

// Request struct
type Request struct {
	client           *httpclient.Client
	minHTTPErrorCode int
}

// HTTPRequest interface
type HTTPRequest interface {
	Do(context context.Context, method, url string, reqBody []byte, headers map[string]string) ([]byte, error)
}

// NewHTTPRequest function
// Request's Constructor
// Returns : *Request
func NewHTTPRequest(retries int, sleepBetweenRetry time.Duration, minHTTPErrorCode int) HTTPRequest {
	// define a maximum jitter interval
	maximumJitterInterval := 5 * time.Millisecond

	// create a backoff
	backoff := heimdall.NewConstantBackoff(sleepBetweenRetry, maximumJitterInterval)

	// create a new retry mechanism with the backoff
	retrier := heimdall.NewRetrier(backoff)

	// set http timeout
	timeout := 10000 * time.Millisecond

	// set http client
	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(retries),
	)

	if minHTTPErrorCode <= 0 {
		minHTTPErrorCode = http.StatusBadRequest
	}

	return &Request{
		client:           client,
		minHTTPErrorCode: minHTTPErrorCode,
	}
}

// Do function, for http client call
func (request *Request) Do(ctx context.Context, method, url string, requestBody []byte, headers map[string]string) (respBody []byte, err error) {
	// set request http
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// set tracer
	trace := tracer.StartTrace(ctx, fmt.Sprintf("HTTP Request: %s %s%s", method, req.URL.Host, req.URL.Path))
	defer func() {
		if err != nil {
			trace.SetError(err)
		}
		trace.Finish()
	}()

	// iterate optional data of headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	trace.InjectHTTPHeader(req)
	tags := trace.Tags()
	tags["http.headers"] = req.Header
	tags["http.method"] = req.Method
	tags["http.url"] = req.URL.String()
	if requestBody != nil {
		tags["request.body"] = string(requestBody)
	}

	// client request
	r, err := request.client.Do(req)
	if err != nil {
		return nil, err
	}
	// close response body
	defer r.Body.Close()

	respBody, err = ioutil.ReadAll(r.Body)

	tags["response.body"] = string(respBody)
	tags["response.code"] = r.StatusCode
	tags["response.status"] = r.Status

	if r.StatusCode >= request.minHTTPErrorCode {
		err = errors.New(r.Status)
		var resp map[string]string
		json.Unmarshal(respBody, &resp)
		if resp["message"] != "" {
			err = errors.New(resp["message"])
		}
		return respBody, err
	}

	return respBody, err
}
