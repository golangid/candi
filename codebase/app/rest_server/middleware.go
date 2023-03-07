package restserver

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golangid/candi/logger"
	"github.com/labstack/echo"
)

// EchoWrapMiddleware wraps `func(http.Handler) http.Handler` into `echo.MiddlewareFunc`
func EchoWrapMiddleware(m func(http.Handler) http.Handler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.SetRequest(r)
				c.Response().Writer = w
				err = next(c)
			})).ServeHTTP(c.Response().Writer, c.Request())

			return
		}
	}
}

// EchoLoggerMiddleware middleware
func EchoLoggerMiddleware(isActive bool, writer io.Writer) echo.MiddlewareFunc {
	bPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			start := time.Now()

			errNext := next(c)
			if errNext != nil {
				c.Error(errNext)
			}

			if !isActive {
				return nil
			}

			req := c.Request()
			res := c.Response()

			stop := time.Now()
			buf := bPool.Get().(*bytes.Buffer)
			buf.Reset()
			defer bPool.Put(buf)

			buf.WriteString(`{"time":"`)
			buf.WriteString(time.Now().Format(time.RFC3339Nano))

			buf.WriteString(`","id":"`)
			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			buf.WriteString(id)

			buf.WriteString(`","remote_ip":"`)
			buf.WriteString(c.RealIP())

			buf.WriteString(`","host":"`)
			buf.WriteString(req.Host)

			buf.WriteString(`","method":"`)
			buf.WriteString(req.Method)

			buf.WriteString(`","uri":"`)
			buf.WriteString(req.RequestURI)

			buf.WriteString(`","user_agent":"`)
			buf.WriteString(req.UserAgent())

			buf.WriteString(`","status":`)
			n := res.Status
			s := logger.GreenColor(n)
			switch {
			case n >= 500:
				s = logger.RedColor(n)
			case n >= 400:
				s = logger.YellowColor(n)
			case n >= 300:
				s = logger.CyanColor(n)
			}
			buf.WriteString(s)

			buf.WriteString(`,"error":"`)
			if errNext != nil {
				buf.WriteString(logger.RedColor(errNext.Error()))
			}

			buf.WriteString(`","latency":`)
			l := stop.Sub(start)
			buf.WriteString(strconv.FormatInt(int64(l), 10))

			buf.WriteString(`,"latency_human":"`)
			buf.WriteString(stop.Sub(start).String())

			buf.WriteString(`","bytes_in":`)
			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}
			buf.WriteString(cl)

			buf.WriteString(`,"bytes_out":`)
			buf.WriteString(strconv.FormatInt(res.Size, 10))

			buf.WriteString("}\n")

			io.Copy(writer, buf)
			return nil
		}
	}
}
