package httpapi

import (
	"log"
	"net/http"
)

func WithRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered method=%s path=%s panic=%v", r.Method, r.URL.Path, recovered)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
