package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"github.com/labstack/echo"
)

// Logger function for writing all request log into console
func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		req := c.Request()
		res := c.Response()

		err := next(c)

		statusCode := res.Status
		if he, ok := err.(*echo.HTTPError); ok {
			statusCode = he.Code
		}
		end := time.Now()

		statusColor := colorForStatus(statusCode)
		methodColor := colorForMethod(req.Method)
		resetColor := helper.Reset

		fmt.Fprintf(os.Stdout, "%s[SERVICE-REST]%s :%s %v | %s %3d %s | %13v | %15s | %s %-7s %s %s\n",
			helper.White, helper.Reset, req.URL.Port(),
			end.Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, resetColor,
			end.Sub(start),
			c.RealIP(),
			methodColor, req.Method, resetColor,
			req.RequestURI,
		)
		return err
	}
}

func colorForStatus(code int) []byte {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return helper.Green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return helper.White
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return helper.Yellow
	default:
		return helper.Red
	}
}

func colorForMethod(method string) []byte {
	switch method {
	case "GET":
		return helper.Blue
	case "POST":
		return helper.Cyan
	case "PUT":
		return helper.Yellow
	case "DELETE":
		return helper.Red
	case "PATCH":
		return helper.Green
	case "HEAD":
		return helper.Magenta
	case "OPTIONS":
		return helper.White
	default:
		return helper.Reset
	}
}
