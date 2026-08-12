//go:build integration

package connectionlimit

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
	"github.com/Rowlyge/kuflow/internal/config"
	middleware "github.com/Rowlyge/kuflow/internal/middleware"
	authservice "github.com/Rowlyge/kuflow/internal/service/auth"
)

func TestConnectionLimitHTTPChain(t *testing.T) {
	cache := authcache.New()

	cache.Replace(map[string]authcache.APIKey{
		"key-1": {
			ID:      1,
			Key:     "key-1",
			Owner:   "integration-test",
			Enabled: true,
		},
	})

	authService := authservice.New(
		authservice.NewValidator(cache),
	)

	connectionLimiter := New(1)

	connectionMiddleware := NewMiddleware(
		connectionLimiter,
		config.AuthConfig{
			APIKeyHeader: "X-API-Key",
		},
	)

	started := make(chan struct{})
	release := make(chan struct{})

	var mu sync.Mutex
	requestCount := 0

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		mu.Lock()
		requestCount++
		currentRequest := requestCount
		mu.Unlock()

		// Только первый запрос удерживает connection slot.
		if currentRequest == 1 {
			close(started)

			<-release
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Default(
		next,
		middleware.NewAuth(authService),
		connectionMiddleware.Wrap,
	)

	// --------------------------------------------------
	// Первый запрос
	// --------------------------------------------------

	firstRequest := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	firstRequest.Header.Set(
		"X-API-Key",
		"key-1",
	)

	firstResponse := httptest.NewRecorder()

	var firstDone sync.WaitGroup
	firstDone.Add(1)

	go func() {
		defer firstDone.Done()

		handler.ServeHTTP(
			firstResponse,
			firstRequest,
		)
	}()

	// Убеждаемся, что первый запрос уже прошёл
	// Auth + ConnectionLimit и удерживает slot.
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first request did not reach handler")
	}

	// --------------------------------------------------
	// Второй запрос с тем же API key
	// --------------------------------------------------

	secondRequest := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	secondRequest.Header.Set(
		"X-API-Key",
		"key-1",
	)

	secondResponse := httptest.NewRecorder()

	handler.ServeHTTP(
		secondResponse,
		secondRequest,
	)

	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected second request status %d, got %d",
			http.StatusTooManyRequests,
			secondResponse.Code,
		)
	}

	// --------------------------------------------------
	// Освобождаем первый запрос
	// --------------------------------------------------

	close(release)

	firstDone.Wait()

	if firstResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected first request status %d, got %d",
			http.StatusOK,
			firstResponse.Code,
		)
	}

	// --------------------------------------------------
	// Третий запрос
	// --------------------------------------------------

	thirdRequest := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	thirdRequest.Header.Set(
		"X-API-Key",
		"key-1",
	)

	thirdResponse := httptest.NewRecorder()

	handler.ServeHTTP(
		thirdResponse,
		thirdRequest,
	)

	if thirdResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected third request status %d, got %d",
			http.StatusOK,
			thirdResponse.Code,
		)
	}

	// После завершения запроса connection slot должен
	// быть полностью освобождён.
	if active := connectionLimiter.Active("key-1"); active != 0 {
		t.Fatalf(
			"expected zero active connections after request, got %d",
			active,
		)
	}
}

func TestConnectionLimitHTTPChainRejectsUnauthorized(t *testing.T) {
	cache := authcache.New()

	authService := authservice.New(
		authservice.NewValidator(cache),
	)

	connectionLimiter := New(1)

	connectionMiddleware := NewMiddleware(
		connectionLimiter,
		config.AuthConfig{
			APIKeyHeader: "X-API-Key",
		},
	)

	called := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Default(
		next,
		middleware.NewAuth(authService),
		connectionMiddleware.Wrap,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"X-API-Key",
		"invalid-key",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}

	if called {
		t.Fatal(
			"next handler should not be called for unauthorized request",
		)
	}

	if active := connectionLimiter.Active("invalid-key"); active != 0 {
		t.Fatalf(
			"expected zero active connections for unauthorized request, got %d",
			active,
		)
	}
}
