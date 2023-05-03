package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
)

const (
	defaultCacheAge = 1 * time.Minute
)

// HTTPCache middleware for cache
func (m *Middleware) HTTPCache(next http.Handler) http.Handler {

	type cacheData struct {
		Body       []byte      `json:"body,omitempty"`
		StatusCode int         `json:"statusCode,omitempty"`
		Header     http.Header `json:"header,omitempty"`
	}

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		trace, ctx := tracer.StartTraceWithContext(req.Context(), "Middleware:HTTPCache")
		defer trace.Finish()

		cacheControl := req.Header.Get(HeaderCacheControl)

		var (
			noCache = m.cache == nil
			maxAge  = m.defaultCacheAge
		)

		for _, headerCacheControl := range strings.Split(cacheControl, ",") {
			key, val, _ := strings.Cut(headerCacheControl, "=")
			switch strings.TrimSpace(key) {
			case "max-age", "s-maxage":
				noCache = false
				age, err := strconv.Atoi(val)
				if err != nil {
					continue
				}
				maxAge = time.Duration(age) * time.Second

			case "no-cache":
				noCache = true

			}
		}

		trace.Log("no-cache", noCache)
		if noCache || maxAge <= 0 {
			next.ServeHTTP(res, req)
			return
		}

		cacheKey := req.URL.String()
		if cacheVal, err := m.cache.Get(ctx, cacheKey); err == nil {

			ttl, _ := m.cache.GetTTL(ctx, cacheKey)
			res.Header().Add(HeaderExpires, time.Now().In(time.UTC).Add(ttl).Format(time.RFC1123))

			var data cacheData
			json.Unmarshal(cacheVal, &data)
			for k := range data.Header {
				res.Header().Set(k, data.Header.Get(k))
			}
			res.Write(data.Body)
			res.WriteHeader(data.StatusCode)
			return
		}

		resBody := &bytes.Buffer{}
		respWriter := wrapper.NewWrapHTTPResponseWriter(resBody, res)

		next.ServeHTTP(respWriter, req)

		if respWriter.StatusCode() <= http.StatusBadRequest {
			m.cache.Set(ctx, req.URL.String(), candihelper.ToBytes(
				cacheData{
					Body:       resBody.Bytes(),
					StatusCode: respWriter.StatusCode(),
					Header:     res.Header(),
				},
			), maxAge)
		}
	})
}
