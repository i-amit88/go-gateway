package httpapi

import (
	"fmt"
	"net/http"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "admin area")
}

func NewRouter(proxy http.Handler, proxyTimeout time.Duration, apiKey string) http.Handler {
	mux := http.NewServeMux()
	protectedAdmin := WithAPIKey(
		apiKey,
		http.HandlerFunc(adminHandler),
	)
	mux.HandleFunc("/health", healthHandler)

	proxyHandler := http.StripPrefix("/proxy", proxy)
	timeoutProxy := withTimeout(
		proxyTimeout,
		proxyHandler,
	)
	mux.Handle("/proxy/", timeoutProxy)
	mux.Handle("/admin", protectedAdmin)

	return mux
}
