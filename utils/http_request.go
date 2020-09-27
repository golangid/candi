package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
)

type (
	// httpRequest struct
	httpRequest struct {
		retries int

		sleepBetweenRetry time.Duration
		client            *http.Client
	}

	// HTTPRequest abstraction
	HTTPRequest interface {
		Do(breakerName, method, url string, body interface{}, headers map[string]string) ([]byte, error)
	}
)

// NewHTTPRequest function
// httpRequest's Constructor
func NewHTTPRequest(retries int, sleepBetweenRetry time.Duration) HTTPRequest {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}

	client := &http.Client{}
	client.Transport = transport
	return &httpRequest{
		retries:           retries,
		sleepBetweenRetry: sleepBetweenRetry,
		client:            client,
	}
}

// Do function, for http client call
func (r *httpRequest) Do(breakerName, method, url string, body interface{}, headers map[string]string) ([]byte, error) {
	hystrix.ConfigureCommand(breakerName, hystrix.CommandConfig{
		Timeout:               int(10 * time.Second),
		MaxConcurrentRequests: 10,
		ErrorPercentThreshold: 25,
	})

	output := make(chan []byte, 1)
	errors := hystrix.Go(breakerName,
		func() error {
			return r.retry(output, method, url, body, headers)
		},
		func(err error) error {
			return err
		})

	select {
	case out := <-output:
		return out, nil
	case err := <-errors:
		return nil, err
	}
}

func (r *httpRequest) retry(output chan []byte, method, url string, body interface{}, headers map[string]string) error {
	ret := retrier.New(retrier.ConstantBackoff(r.retries, r.sleepBetweenRetry), nil)
	attempt := 0
	err := ret.Run(func() error {
		attempt++
		var req *http.Request

		if body != nil {
			typeInput := reflect.TypeOf(body)
			typeReader := reflect.TypeOf((*io.Reader)(nil)).Elem()
			if typeInput.Implements(typeReader) {
				reader := body.(io.Reader)
				req, _ = http.NewRequest(method, url, reader)
			} else {
				payload, _ := json.Marshal(body)
				buf := bytes.NewBuffer(payload)
				req, _ = http.NewRequest(method, url, buf)
			}
		} else {
			req, _ = http.NewRequest(method, url, nil)
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := r.client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode < 499 {
				responseBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				output <- responseBody
				return nil
			}
			err = fmt.Errorf("Status was %d", resp.StatusCode)
		}
		return err
	})

	return err
}
