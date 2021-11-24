package httprouter

import (
	"net/http"
	"time"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Pusher

	StatusCode() int
	Size() int
	Latency() time.Duration
}

type responseWriter struct {
	delegate   http.ResponseWriter
	statusCode int
	size       int
	start      time.Time
}

type responseWriterFlusher struct {
	responseWriter
}

func NewResponseWriter(delegate http.ResponseWriter) ResponseWriter {
	rw := responseWriter{
		delegate: delegate,
		start:    time.Now(),
	}

	if _, ok := delegate.(http.Flusher); ok {
		return &responseWriterFlusher{
			responseWriter: rw,
		}
	}

	return &rw
}

func (r *responseWriter) Header() http.Header {
	return r.delegate.Header()
}

// Write implements http.ResponseWriter
func (r *responseWriter) Write(data []byte) (int, error) {
	n, err := r.delegate.Write(data)
	r.size += n

	return n, err
}

// WriteHeader implements http.ResponseWriter
func (r *responseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.delegate.WriteHeader(statusCode)
}

// Push implements http.Pusher
func (r *responseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := r.delegate.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}

	return pusher.Push(target, opts)
}

// Size returns total number of bytes written to response
func (r *responseWriter) Size() int {
	return r.size
}

// StatusCode returns http status code set by WriteHeader
func (r *responseWriter) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}

	return r.statusCode
}

// Latency records time since ResponseWriter was created
func (r *responseWriter) Latency() time.Duration {
	return time.Since(r.start)
}

// Flush implements http.Flusher
func (r *responseWriterFlusher) Flush() {
	flusher := r.delegate.(http.Flusher)
	flusher.Flush()
}
