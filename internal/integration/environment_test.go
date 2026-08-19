package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
)

func TestEnvironmentHealthEndpoint(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	resp, err := env.Client().Get(
		env.URL() + "/health",
	)
	if err != nil {
		t.Fatalf("GET /health error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"GET /health status = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	if requestID := resp.Header.Get("X-Request-ID"); requestID == "" {
		t.Fatal("GET /health did not return X-Request-ID")
	}

	var body map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf(
			"decode /health response: %v",
			err,
		)
	}

	if body["status"] != "ok" {
		t.Fatalf(
			"GET /health response status = %v, want %q",
			body["status"],
			"ok",
		)
	}
}

func TestEnvironmentProxyPipeline(t *testing.T) {
	var requests atomic.Int64

	upstream := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requests.Add(1)

			if got := r.Header.Get("X-Forwarded-Proto"); got != "http" {
				t.Errorf(
					"X-Forwarded-Proto = %q, want %q",
					got,
					"http",
				)
			}

			if got := r.Header.Get("X-Forwarded-Host"); got == "" {
				t.Error("X-Forwarded-Host is empty")
			}

			w.Header().Set("X-Upstream", "integration")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("upstream response"))
		}),
	)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test?value=42",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-valid-key",
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(
			"proxy response status = %d, want %d",
			resp.StatusCode,
			http.StatusCreated,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read proxy response: %v", err)
	}

	if string(body) != "upstream response" {
		t.Fatalf(
			"proxy response body = %q, want %q",
			string(body),
			"upstream response",
		)
	}

	if got := resp.Header.Get("X-KuFlow"); got != "true" {
		t.Fatalf(
			"X-KuFlow = %q, want %q",
			got,
			"true",
		)
	}

	if requestID := resp.Header.Get("X-Request-ID"); requestID == "" {
		t.Fatal("proxy response did not contain X-Request-ID")
	}

	if got := resp.Header.Get("X-Upstream"); got != "integration" {
		t.Fatalf(
			"X-Upstream = %q, want %q",
			got,
			"integration",
		)
	}

	if got := requests.Load(); got != 1 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			1,
		)
	}
}

func newAuthTestUpstream(
	t *testing.T,
	requests *atomic.Int64,
) *httptest.Server {
	t.Helper()

	return httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requests.Add(1)

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("upstream response"))
		}),
	)
}

func newValidAPIKey() authcache.APIKey {
	return authcache.APIKey{
		ID:      1,
		Key:     "integration-valid-key",
		Owner:   "integration-test",
		Enabled: true,
	}
}

func TestEnvironmentAuthenticationValidKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-valid-key",
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusCreated,
		)
	}

	if got := requests.Load(); got != 1 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			1,
		)
	}
}

func TestEnvironmentAuthenticationMissingKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	resp, err := env.Client().Get(
		env.URL() + "/proxy-test",
	)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}

func TestEnvironmentAuthenticationInvalidKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-invalid-key",
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}

func TestEnvironmentAuthenticationDisabledKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	key := newValidAPIKey()
	key.Enabled = false

	env.SetAPIKeys(key)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		key.Key,
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}

func TestEnvironmentAuthenticationExpiredKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	expiredAt := time.Now().Add(-time.Minute)

	key := newValidAPIKey()
	key.ExpiresAt = &expiredAt

	env.SetAPIKeys(key)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		key.Key,
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}
