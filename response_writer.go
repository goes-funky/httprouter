package httprouter

import (
	"net/http"
	"time"
)

type ResponseWriter struct {
	http.ResponseWriter
	start      time.Time
	statusCode int
	size       int
}

func NewResponseWriter(delegate http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: delegate,
		start:          time.Now(),
	}
}

// Write implements http.ResponseWriter
func (r *ResponseWriter) Write(data []byte) (int, error) {
	n, err := r.ResponseWriter.Write(data)
	r.size += n

	return n, err
}

// WriteHeader implements http.ResponseWriter
func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Size returns total number of bytes written to response
func (r *ResponseWriter) Size() int {
	return r.size
}

func (r *ResponseWriter) StatusCode() int {
	return r.statusCode
}

func (r *ResponseWriter) Latency() time.Duration {
	return time.Since(r.start)
}
