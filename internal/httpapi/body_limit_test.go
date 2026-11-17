package httpapi

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithBodyLimitAllowsBodyWithinLimit(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("expected body to be readable, got error: %v", err)
		}

		if string(body) != "hello" {
			t.Fatalf("expected body %q, got %q", "hello", string(body))
		}

		w.WriteHeader(http.StatusNoContent)
	})

	handler := WithBodyLimit(10, next)
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader("hello"))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
}

func TestWithBodyLimitRejectsBodyOverLimit(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Fatal("expected body read to fail")
		}

		var maxBytesError *http.MaxBytesError
		if !errors.As(err, &maxBytesError) {
			t.Fatalf("expected MaxBytesError, got %T: %v", err, err)
		}

		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
	})

	handler := WithBodyLimit(5, next)
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader("too-large"))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", res.Code)
	}
}
