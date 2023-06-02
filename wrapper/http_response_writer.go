package wrapper

import (
	"bytes"
	"net/http"
)

// WrapHTTPResponseWriter wrapper
type WrapHTTPResponseWriter struct {
	statusCode     int
	buff           *bytes.Buffer
	maxWriteSize   int
	limitWriteSize bool
	contentLength  int
	http.ResponseWriter
}

// NewWrapHTTPResponseWriter init new wrapper for http response writter
func NewWrapHTTPResponseWriter(responseBuff *bytes.Buffer, httpResponseWriter http.ResponseWriter) *WrapHTTPResponseWriter {
	return &WrapHTTPResponseWriter{
		statusCode: http.StatusOK, buff: responseBuff, ResponseWriter: httpResponseWriter,
	}
}

// SetMaxWriteSize set max write size to buffer
func (w *WrapHTTPResponseWriter) SetMaxWriteSize(max int) {
	w.maxWriteSize = max
	w.limitWriteSize = true
}

// GetContentLength get response content length
func (w *WrapHTTPResponseWriter) GetContentLength() int {
	return w.contentLength
}

// GetContent get response content
func (w *WrapHTTPResponseWriter) GetContent() []byte {
	return w.buff.Bytes()
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
	n, err := w.ResponseWriter.Write(data)
	w.contentLength += n
	if !w.limitWriteSize || w.contentLength < w.maxWriteSize {
		w.buff.Write(data)
	}
	return n, err
}

// WriteHeader method
func (w *WrapHTTPResponseWriter) WriteHeader(statusCode int) {
	// Store the status code
	w.statusCode = statusCode

	// Write the status code onward.
	w.ResponseWriter.WriteHeader(statusCode)
}
