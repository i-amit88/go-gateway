package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
