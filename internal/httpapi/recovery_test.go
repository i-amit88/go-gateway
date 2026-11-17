package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRecoveryReturnsInternalServerErrorAfterPanic(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := WithRecovery(next)
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), "internal server error") {
		t.Fatalf("expected internal server error response, got %q", res.Body.String())
	}
}

func TestWithRecoveryPassesThroughNormalRequest(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	})

	handler := WithRecovery(next)
	req := httptest.NewRequest(http.MethodGet, "/normal", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", res.Code)
	}

	if res.Body.String() != "accepted" {
		t.Fatalf("expected body %q, got %q", "accepted", res.Body.String())
	}
}
