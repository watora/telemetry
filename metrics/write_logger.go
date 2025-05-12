package metrics

import "net/http"

type writeLogger struct {
	http.ResponseWriter
	statusCode int
}

func (w *writeLogger) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
