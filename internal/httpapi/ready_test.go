package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadyHandlerReturnsReadyWhenCheckSucceeds(t *testing.T) {
	checkCalled := false
	check := func(ctx context.Context) error {
		checkCalled = true
		return nil
	}

	handler := readyHandler(check)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if !checkCalled {
		t.Fatal("expected readiness check to be called")
	}

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if res.Body.String() != "READY\n" {
		t.Fatalf("expected body %q, got %q", "READY\n", res.Body.String())
	}
}

func TestReadyHandlerReturnsNotReadyWhenCheckFails(t *testing.T) {
	check := func(ctx context.Context) error {
		return errors.New("upstream unavailable")
	}

	handler := readyHandler(check)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), "not ready") {
		t.Fatalf("expected not-ready response, got %q", res.Body.String())
	}
}

func TestReadyHandlerPassesRequestCancellationToCheck(t *testing.T) {
	requestContext, cancelRequest := context.WithCancel(context.Background())
	cancelRequest()

	check := func(ctx context.Context) error {
		if ctx.Err() == nil {
			t.Fatal("expected readiness context to be canceled")
		}
		return ctx.Err()
	}

	handler := readyHandler(check)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil).WithContext(requestContext)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", res.Code)
	}
}
