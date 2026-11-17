package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleProxyErrorReturnsGatewayTimeout(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/proxy/slow", nil)
	res := httptest.NewRecorder()

	handleProxyError(res, req, context.DeadlineExceeded)

	if res.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d", res.Code)
	}

	if res.Body.String() != "upstream timeout\n" {
		t.Fatalf("unexpected response body %q", res.Body.String())
	}
}

func TestHandleProxyErrorReturnsBadGateway(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/proxy/hello", nil)
	res := httptest.NewRecorder()

	handleProxyError(res, req, errors.New("connection refused"))

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", res.Code)
	}

	if res.Body.String() != "upstream unavailable\n" {
		t.Fatalf("unexpected response body %q", res.Body.String())
	}
}

func TestCheckUpstreamSucceedsWhenHealthEndpointIsHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("expected health path, got %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := checkUpstream(context.Background(), server.Client(), server.URL); err != nil {
		t.Fatalf("expected healthy upstream check to pass, got %v", err)
	}
}

func TestCheckUpstreamReturnsErrorWhenServerIsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close()

	if err := checkUpstream(context.Background(), http.DefaultClient, serverURL); err == nil {
		t.Fatal("expected unavailable upstream check to fail")
	}
}

func TestCheckUpstreamReturnsErrorWhenContextTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := checkUpstream(ctx, server.Client(), server.URL)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded error, got %v", err)
	}
}
