package httpapi

import (
	"log"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(body)
}

func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseWriter := &statusWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		defer func() {
			log.Printf(
				"method=%s path=%s status=%d request_id=%s duration=%s",
				r.Method,
				r.URL.Path,
				responseWriter.statusCode,
				RequestIDFromContext(r.Context()),
				time.Since(start),
			)
		}()

		next.ServeHTTP(responseWriter, r) // after this every writer is this one own warapperd
	})
}
