package main

const (
	templateSDK = `// {{.Header}} DO NOT EDIT.

package sdk

import (
	"sync"

	// @candi:serviceImport
)

// Option func type
type Option func(*sdkInstance)

var (
	sdk  SDK
	once sync.Once
)

// SetGlobalSDK constructor with each sdk service option.
func SetGlobalSDK(opts ...Option) {
	s := new(sdkInstance)
	for _, o := range opts {
		o(s)
	}
	once.Do(func() {
		sdk = s
	})
}

// GetSDK get global sdk instance
func GetSDK() SDK {
	return sdk
}

// @candi:construct

// SDK instance abstraction
type SDK interface {
	// @candi:serviceMethod
}

// sdkInstance implementation
type sdkInstance struct {
	// @candi:serviceField
}

// @candi:instanceMethod
`

	templateSDKServiceAbstraction = `package {{lower (clean $.ServiceName)}}

// {{upper (clean $.ServiceName)}} client abstract interface
type {{upper (clean $.ServiceName)}} interface {
	// Add service client method
}
`

	templateSDKServiceGRPC = `package {{lower (clean $.ServiceName)}}

import (
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

type {{lower (clean $.ServiceName)}}GRPCImpl struct {
	host    string
	authKey string
	conn    *grpc.ClientConn
}

// New{{upper (clean $.ServiceName)}}ServiceGRPC constructor
func New{{upper (clean $.ServiceName)}}ServiceGRPC(host string, authKey string) {{upper (clean $.ServiceName)}} {

	if u, _ := url.Parse(host); u.Host != "" {
		host = u.Host
	}
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  50 * time.Millisecond,
			Multiplier: 5,
			MaxDelay:   50 * time.Millisecond,
		},
		MinConnectTimeout: 1 * time.Second,
	}))
	if err != nil {
		panic(err)
	}

	return &{{lower (clean $.ServiceName)}}GRPCImpl{
		host:    host,
		authKey: authKey,
		conn:    conn,
	}
}
`

	templateSDKServiceREST = `package {{lower (clean $.ServiceName)}}

import (
	"net/http"
	"time"

	"{{.LibraryName}}/candiutils"
)

type {{lower (clean $.ServiceName)}}RESTImpl struct {
	host    string
	authKey string
	httpReq candiutils.HTTPRequest
}

// New{{upper (clean $.ServiceName)}}ServiceREST constructor
func New{{upper (clean $.ServiceName)}}ServiceREST(host string, authKey string) {{upper (clean $.ServiceName)}} {

	return &{{lower (clean $.ServiceName)}}RESTImpl{
		host:    host,
		authKey: authKey,
		httpReq: candiutils.NewHTTPRequest(
			candiutils.HTTPRequestSetRetries(5),
			candiutils.HTTPRequestSetSleepBetweenRetry(500*time.Millisecond),
			candiutils.HTTPRequestSetHTTPErrorCodeThreshold(http.StatusBadRequest),
			candiutils.HTTPRequestSetBreakerName("{{lower (clean $.ServiceName)}}"),
		),
	}
}
`
)
