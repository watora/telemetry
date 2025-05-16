package metrics

import "net/http"

type WriteLogger struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WriteLogger) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
