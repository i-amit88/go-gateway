package httpapi

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestIDAddsHeaderAndContextValue(t *testing.T) {
	var contextRequestID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextRequestID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	handler := WithRequestID(next)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	headerRequestID := res.Header().Get(RequestIDHeader)
	if headerRequestID == "" {
		t.Fatal("expected request ID response header")
	}

	if contextRequestID == "" {
		t.Fatal("expected request ID in request context")
	}

	if contextRequestID != headerRequestID {
		t.Fatalf("expected context request ID %q, got header request ID %q", contextRequestID, headerRequestID)
	}

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
}

func TestRequestIDFromContextReturnsEmptyWhenMissing(t *testing.T) {
	if requestID := RequestIDFromContext(context.Background()); requestID != "" {
		t.Fatalf("expected empty request ID, got %q", requestID)
	}
}

func TestWithLoggingLogsRequestDetails(t *testing.T) {
	var logs bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	handler := WithLogging(next)
	req := httptest.NewRequest(http.MethodPost, "/echo", nil)
	ctx := context.WithValue(req.Context(), requestIDKey, "test-request-id")
	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	output := logs.String()
	for _, expected := range []string{
		"method=POST",
		"path=/echo",
		"status=202",
		"request_id=test-request-id",
		"duration=",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected log output to contain %q, got %q", expected, output)
		}
	}

	if res.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", res.Code)
	}
}

func TestWithLoggingRecordsImplicitOKStatus(t *testing.T) {
	var logs bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	handler := WithLogging(next)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if !strings.Contains(logs.String(), "status=200") {
		t.Fatalf("expected log output to contain status=200, got %q", logs.String())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
}
