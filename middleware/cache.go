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
	// DefaultCacheAge const
	DefaultCacheAge = 1 * time.Minute
)

// HTTPCache middleware for cache
func (m *Middleware) HTTPCache(next http.Handler) http.Handler {
	type cacheData struct {
		Body   []byte      `json:"body,omitempty"`
		Header http.Header `json:"header,omitempty"`
	}

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if m.cache == nil {
			next.ServeHTTP(res, req)
			return
		}

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

		if noCache || maxAge <= 0 {
			next.ServeHTTP(res, req)
			return
		}

		trace, ctx := tracer.StartTraceWithContext(req.Context(), "Middleware:HTTPCache")
		defer trace.Finish()

		cacheKey := req.Method + ":" + strings.TrimSuffix(req.URL.String(), "/")
		trace.SetTag("key", cacheKey)
		if cacheVal, err := m.cache.Get(ctx, cacheKey); err == nil {
			if ttl, err := m.cache.GetTTL(ctx, cacheKey); err == nil {
				res.Header().Add(HeaderExpires, time.Now().In(time.UTC).Add(ttl).Format(time.RFC1123))
			}

			var data cacheData
			if err := json.Unmarshal(cacheVal, &data); err != nil {
				m.cache.Delete(ctx, cacheKey)
				next.ServeHTTP(res, req)
				return
			}

			for k := range data.Header {
				res.Header().Set(k, data.Header.Get(k))
			}
			res.Write(data.Body)
			return
		}

		resBody := &bytes.Buffer{}
		respWriter := wrapper.NewWrapHTTPResponseWriter(resBody, res)

		next.ServeHTTP(respWriter, req)

		if respWriter.StatusCode() < http.StatusBadRequest {
			m.cache.Set(ctx, cacheKey, candihelper.ToBytes(
				cacheData{
					Body:   resBody.Bytes(),
					Header: res.Header(),
				},
			), maxAge)
		}
	})
}
