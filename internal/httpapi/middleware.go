package httpapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request_id"

const RequestIDHeader = "X-Request-ID"

func newRequestID() string {
	return uuid.New().String()
}

// request id middleware
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()

		ctx := context.WithValue(
			r.Context(),
			requestIDKey,
			requestID,
		)

		w.Header().Set(
			RequestIDHeader,
			requestID,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return requestID
}
