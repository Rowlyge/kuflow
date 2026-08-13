//go:build integration

package proxy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Rowlyge/kuflow/internal/breaker"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

type integrationBalancer struct {
	upstream *upstream.Upstream
	err      error
}

func (b *integrationBalancer) Next() (*upstream.Upstream, error) {
	if b.err != nil {
		return nil, b.err
	}

	return b.upstream, nil
}

func newIntegrationUpstream(
	t *testing.T,
	target string,
	cfg breaker.Config,
) *upstream.Upstream {
	t.Helper()

	serverURL := target

	parsedURL, err := upstreamURL(serverURL)
	if err != nil {
		t.Fatalf("failed to parse upstream URL: %v", err)
	}

	up := &upstream.Upstream{
		Name:    "integration-upstream",
		URL:     parsedURL,
		Breaker: breaker.New(cfg),
	}

	up.SetAlive(true)

	return up
}

func upstreamURL(
	target string,
) (*url.URL, error) {
	return url.Parse(target)
}

func integrationProxyConfig() config.ProxyConfig {
	return config.ProxyConfig{
		DialTimeout:           time.Second,
		ResponseHeaderTimeout: time.Second,
		FlushInterval:         0,
	}
}

func TestEngineCircuitBreakerIntegration(t *testing.T) {
	var requests atomic.Int64

	upstreamServer := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requests.Add(1)

			http.Error(
				w,
				"upstream failure",
				http.StatusInternalServerError,
			)
		}),
	)

	defer upstreamServer.Close()

	breakerConfig := breaker.Config{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      20 * time.Millisecond,
	}

	up := newIntegrationUpstream(
		t,
		upstreamServer.URL,
		breakerConfig,
	)

	b := &integrationBalancer{
		upstream: up,
	}

	engine, err := NewEngine(
		b,
		integrationProxyConfig(),
	)

	if err != nil {
		t.Fatalf(
			"failed to create proxy engine: %v",
			err,
		)
	}

	t.Run(
		"first request reaches upstream",
		func(t *testing.T) {

			req := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)

			rec := httptest.NewRecorder()

			engine.ServeHTTP(
				rec,
				req,
			)

			if rec.Code != http.StatusInternalServerError {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusInternalServerError,
					rec.Code,
				)
			}

			if requests.Load() != 1 {
				t.Fatalf(
					"expected 1 upstream request, got %d",
					requests.Load(),
				)
			}

			if up.Breaker.State() != breaker.Open {
				t.Fatalf(
					"expected breaker state Open, got %v",
					up.Breaker.State(),
				)
			}
		},
	)

	t.Run(
		"open breaker prevents upstream request",
		func(t *testing.T) {

			req := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)

			rec := httptest.NewRecorder()

			engine.ServeHTTP(
				rec,
				req,
			)

			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusServiceUnavailable,
					rec.Code,
				)
			}

			if requests.Load() != 1 {
				t.Fatalf(
					"expected upstream request count to remain 1, got %d",
					requests.Load(),
				)
			}
		},
	)

	t.Run(
		"half open allows test request",
		func(t *testing.T) {

			time.Sleep(
				breakerConfig.OpenTimeout +
					10*time.Millisecond,
			)

			if up.Breaker.State() != breaker.Open {
				t.Fatalf(
					"expected breaker to remain Open before Allow, got %v",
					up.Breaker.State(),
				)
			}

			req := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)

			rec := httptest.NewRecorder()

			engine.ServeHTTP(
				rec,
				req,
			)

			if rec.Code != http.StatusInternalServerError {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusInternalServerError,
					rec.Code,
				)
			}

			if requests.Load() != 2 {
				t.Fatalf(
					"expected 2 upstream requests, got %d",
					requests.Load(),
				)
			}

			if up.Breaker.State() != breaker.Open {
				t.Fatalf(
					"expected breaker to reopen after failed half-open request, got %v",
					up.Breaker.State(),
				)
			}
		},
	)
}

func TestEngineUnavailableUpstreamDoesNotRecordFailure(
	t *testing.T,
) {
	breakerConfig := breaker.Config{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      time.Minute,
	}

	up := &upstream.Upstream{
		Name: "unavailable-upstream",
		Breaker: breaker.New(
			breakerConfig,
		),
	}

	up.SetAlive(false)

	b := &integrationBalancer{
		err: errors.New("no available upstream"),
	}

	engine, err := NewEngine(
		b,
		integrationProxyConfig(),
	)

	if err != nil {
		t.Fatalf(
			"failed to create proxy engine: %v",
			err,
		)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	engine.ServeHTTP(
		rec,
		req,
	)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}

	if up.Breaker.State() != breaker.Closed {
		t.Fatalf(
			"expected breaker to remain Closed, got %v",
			up.Breaker.State(),
		)
	}
}
