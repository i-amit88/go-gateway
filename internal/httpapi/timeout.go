package httpapi

import (
	"context"
	"net/http"
	"time"
)

// using the middleware pattern
func withTimeout(
	duration time.Duration,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), duration)
		defer cancel()

		requestWithTimeout := r.WithContext(ctx)

		// call nxt function handler
		next.ServeHTTP(w, requestWithTimeout)
	})
}
