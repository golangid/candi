package wrapper

import (
	"io"
	"net/http"
)

// WrapHTTPResponseWriter wrapper
type WrapHTTPResponseWriter struct {
	statusCode int
	io.Writer
	http.ResponseWriter
}

// NewWrapHTTPResponseWriter init new wrapper for http response writter
func NewWrapHTTPResponseWriter(w io.Writer, httpResponseWriter http.ResponseWriter) *WrapHTTPResponseWriter {
	// Default the status code to 200
	return &WrapHTTPResponseWriter{statusCode: 200, Writer: w, ResponseWriter: httpResponseWriter}
}

// StatusCode give a way to get the Code
func (w *WrapHTTPResponseWriter) StatusCode() int {
	return w.statusCode
}

// Header Satisfy the http.ResponseWriter interface
func (w *WrapHTTPResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *WrapHTTPResponseWriter) Write(data []byte) (int, error) {
	// Store response body to writer
	return w.Writer.Write(data)
}

// WriteHeader method
func (w *WrapHTTPResponseWriter) WriteHeader(statusCode int) {
	// Store the status code
	w.statusCode = statusCode

	// Write the status code onward.
	w.ResponseWriter.WriteHeader(statusCode)
}
