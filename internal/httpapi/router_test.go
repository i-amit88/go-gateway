package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"
)

func testReadyCheck(ctx context.Context) error {
	return nil
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	healthHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if res.Body.String() != "OK\n" {
		t.Fatalf("expected body %q, got %q", "OK\n", res.Body.String())
	}
}

func TestLiveHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	res := httptest.NewRecorder()

	liveHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if res.Body.String() != "OK\n" {
		t.Fatalf("expected body %q, got %q", "OK\n", res.Body.String())
	}
}

func TestLiveRouteIsPublic(t *testing.T) {
	router := NewRouter(http.NotFoundHandler(), time.Second, "test-key", testReadyCheck)
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if res.Body.String() != "OK\n" {
		t.Fatalf("expected body %q, got %q", "OK\n", res.Body.String())
	}
}

func TestRouterProxiesWithoutProxyPrefix(t *testing.T) {
	var receivedPath string

	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		_, _ = w.Write([]byte("response from upstream"))
	})

	router := NewRouter(upstream, 2*time.Second, "test-key", testReadyCheck)
	req := httptest.NewRequest(http.MethodGet, "/proxy/hello", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if receivedPath != "/hello" {
		t.Fatalf("expected upstream path /hello, got %q", receivedPath)
	}

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if res.Body.String() != "response from upstream" {
		t.Fatalf("unexpected response body %q", res.Body.String())
	}
}

func TestAdminRouteRejectsMissingAPIKey(t *testing.T) {
	router := NewRouter(http.NotFoundHandler(), time.Second, "test-key", testReadyCheck)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestAdminRouteRejectsInvalidAPIKey(t *testing.T) {
	router := NewRouter(http.NotFoundHandler(), time.Second, "test-key", testReadyCheck)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set(APIKeyHeader, "wrong-key")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestAdminRouteAcceptsValidAPIKey(t *testing.T) {
	router := NewRouter(http.NotFoundHandler(), time.Second, "test-key", testReadyCheck)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set(APIKeyHeader, "test-key")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if res.Body.String() != "admin area\n" {
		t.Fatalf("unexpected response body %q", res.Body.String())
	}
}

func TestHealthRouteRemainsPublic(t *testing.T) {
	router := NewRouter(http.NotFoundHandler(), time.Second, "test-key", testReadyCheck)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
}

func TestRouterReturnsBadGatewayWhenUpstreamUnavailable(t *testing.T) {
	target, err := url.Parse("http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}

	router := NewRouter(proxy, 2*time.Second, "test-key", testReadyCheck)
	req := httptest.NewRequest(http.MethodGet, "/proxy/hello", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", res.Code)
	}

	if res.Body.String() != "upstream unavailable\n" {
		t.Fatalf("unexpected response body %q", res.Body.String())
	}
}

func TestWithTimeoutCancelsNextHandler(t *testing.T) {
	reachedNext := make(chan struct{})
	observedCancellation := make(chan struct{})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(reachedNext)
		<-r.Context().Done()
		close(observedCancellation)
	})

	handler := withTimeout(20*time.Millisecond, next)
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	res := httptest.NewRecorder()
	done := make(chan struct{})

	go func() {
		handler.ServeHTTP(res, req)
		close(done)
	}()

	select {
	case <-reachedNext:
	case <-time.After(time.Second):
		t.Fatal("next handler was not called")
	}

	select {
	case <-observedCancellation:
	case <-time.After(time.Second):
		t.Fatal("next handler did not observe context cancellation")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout handler did not finish")
	}
}
