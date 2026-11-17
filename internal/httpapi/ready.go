package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func readyHandler(check func(context.Context) error) http.Handler { // check is a function as a input
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)

		defer cancel()

		if err := check(ctx); err != nil {
			http.Error(
				w,
				"not ready",
				http.StatusServiceUnavailable,
			)
			return
		}

		fmt.Fprintln(w, "READY")

	})
}
