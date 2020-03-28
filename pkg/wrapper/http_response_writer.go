package wrapper

import "net/http"

type WrapResponseWriter struct {
	StatusCode int
	http.ResponseWriter
}

func NewWrapResponseWriter(res http.ResponseWriter) *WrapResponseWriter {
	// Default the status code to 200
	return &WrapResponseWriter{200, res}
}

// Give a way to get the status
func (w WrapResponseWriter) Status() int {
	return w.StatusCode
}

// Satisfy the http.ResponseWriter interface
func (w WrapResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w WrapResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (w WrapResponseWriter) WriteHeader(statusCode int) {
	// Store the status code
	w.StatusCode = statusCode

	// Write the status code onward.
	w.ResponseWriter.WriteHeader(statusCode)
}
