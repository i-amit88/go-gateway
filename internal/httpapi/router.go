package httpapi

import (
	"fmt"
	"net/http"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func NewRouter(proxy http.Handler, proxyTimeout time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	proxyHandler := http.StripPrefix("/proxy", proxy)
	timeoutProxy := withTimeout(
		proxyTimeout,
		proxyHandler,
	)
	mux.Handle("/proxy/", timeoutProxy)

	return mux
}
