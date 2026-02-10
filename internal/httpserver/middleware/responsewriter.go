package middleware

import (
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	Status int
	Bytes  int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		Status:         http.StatusOK,
	}
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.Bytes += n
	return n, err
}

func (w *ResponseWriter) StatusCode() int {
	return w.Status
}

func (w *ResponseWriter) BytesWritten() int {
	return w.Bytes
}
