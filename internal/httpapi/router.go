package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"example.com/go-gateway/internal/session"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "admin area")
}

func liveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func NewRouter(
	proxy http.Handler,
	proxyTimeout time.Duration,
	apiKey string,
	readyCheck func(context.Context) error,
	sessionManager session.SessionManager,
) http.Handler {
	mux := http.NewServeMux()
	protectedAdmin := WithAPIKey(
		apiKey,
		http.HandlerFunc(adminHandler),
	)
	mux.HandleFunc("/health", healthHandler)
	proxyHandler := http.StripPrefix("/proxy", proxy)
	sessions := session.NewSessionHandler(sessionManager)
	timeoutProxy := withTimeout(
		proxyTimeout,
		proxyHandler,
	)
	mux.Handle("/proxy/", timeoutProxy)
	mux.Handle("/admin", protectedAdmin)
	mux.HandleFunc("/live", liveHandler)
	mux.Handle(
		"/ready",
		readyHandler(readyCheck),
	)
	mux.HandleFunc("POST /sessions", sessions.Create)
	mux.HandleFunc("GET /sessions/{id}", sessions.Get)
	mux.HandleFunc("DELETE /sessions/{id}", sessions.Delete)

	return mux
}
