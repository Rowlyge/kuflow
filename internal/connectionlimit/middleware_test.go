package connectionlimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rowlyge/kuflow/internal/config"
)

func newTestMiddleware(limit int) *Middleware {
	limiter := New(limit)

	return NewMiddleware(
		limiter,
		config.AuthConfig{
			APIKeyHeader: "X-API-Key",
		},
	)
}

func TestMiddlewareAllowsRequest(t *testing.T) {
	middleware := newTestMiddleware(1)

	called := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Wrap(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"X-API-Key",
		"key-1",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if !called {
		t.Fatal("next handler was not called")
	}
}

func TestMiddlewareRejectsWhenLimitExceeded(t *testing.T) {
	middleware := newTestMiddleware(1)

	limiter := middleware.limiter

	if !limiter.Acquire("key-1") {
		t.Fatal("failed to occupy connection slot")
	}

	defer limiter.Release("key-1")

	called := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Wrap(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"X-API-Key",
		"key-1",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			rec.Code,
		)
	}

	if called {
		t.Fatal("next handler should not be called when limit is exceeded")
	}
}

func TestMiddlewareReleasesConnection(t *testing.T) {
	middleware := newTestMiddleware(1)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Wrap(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"X-API-Key",
		"key-1",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if active := middleware.limiter.Active("key-1"); active != 0 {
		t.Fatalf(
			"expected connection to be released, got %d active connections",
			active,
		)
	}

	if !middleware.limiter.Acquire("key-1") {
		t.Fatal("connection slot should be available after request completion")
	}

	middleware.limiter.Release("key-1")
}

func TestMiddlewareDifferentAPIKeys(t *testing.T) {
	middleware := newTestMiddleware(1)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Wrap(next)

	req1 := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req1.Header.Set(
		"X-API-Key",
		"key-1",
	)

	req2 := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req2.Header.Set(
		"X-API-Key",
		"key-2",
	)

	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)
	handler.ServeHTTP(rec2, req2)

	if rec1.Code != http.StatusOK {
		t.Fatalf(
			"key-1: expected status %d, got %d",
			http.StatusOK,
			rec1.Code,
		)
	}

	if rec2.Code != http.StatusOK {
		t.Fatalf(
			"key-2: expected status %d, got %d",
			http.StatusOK,
			rec2.Code,
		)
	}
}

func TestMiddlewareUsesConfiguredHeader(t *testing.T) {
	limiter := New(1)

	middleware := NewMiddleware(
		limiter,
		config.AuthConfig{
			APIKeyHeader: "X-Custom-Key",
		},
	)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Wrap(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"X-Custom-Key",
		"key-1",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
