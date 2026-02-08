package middleware

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	Status int
	Bytes  int
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	if w.Status == 0 {
		w.Status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.Bytes += n
	return n, err
}

// StatusCode returns the HTTP status code written by the handler.
func (w *ResponseWriter) StatusCode() int {
	return w.Status
}

// BytesWritten returns the number of bytes written to the response.
func (w *ResponseWriter) BytesWritten() int {
	return w.Bytes
}
