package httpapi

import (
	"crypto/subtle"
	"net/http"
)

const APIKeyHeader = "X-API-Key"

func WithAPIKey(expectedKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedKey := r.Header.Get("X-API-Key")

		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
			http.Error(
				w,
				"unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		next.ServeHTTP(w,r)
	})
}
