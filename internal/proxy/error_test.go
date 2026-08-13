package proxy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Rowlyge/kuflow/internal/breaker"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

func newTestUpstream(
	name string,
	cfg breaker.Config,
) *upstream.Upstream {
	return &upstream.Upstream{
		Name: name,
		Breaker: breaker.New(
			cfg,
		),
	}
}

func TestErrorHandlerSelectedUpstreamRecordsFailure(t *testing.T) {
	cfg := breaker.Config{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      time.Minute,
	}

	u := newTestUpstream(
		"test-upstream",
		cfg,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req = req.WithContext(
		IntoContext(
			req.Context(),
			u,
		),
	)

	rec := httptest.NewRecorder()

	handler := newErrorHandler()

	handler(
		rec,
		req,
		errors.New("upstream connection failed"),
	)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadGateway,
			rec.Code,
		)
	}

	if state := u.Breaker.State(); state != breaker.Open {
		t.Fatalf(
			"expected breaker state Open, got %v",
			state,
		)
	}
}

func TestErrorHandlerUnavailableUpstreamDoesNotRecordFailure(
	t *testing.T,
) {
	cfg := breaker.Config{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      time.Minute,
	}

	u := newTestUpstream(
		"test-upstream",
		cfg,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req = req.WithContext(
		MarkUpstreamUnavailable(
			req.Context(),
		),
	)

	rec := httptest.NewRecorder()

	handler := newErrorHandler()

	handler(
		rec,
		req,
		errors.New("no upstream available"),
	)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}

	if state := u.Breaker.State(); state != breaker.Closed {
		t.Fatalf(
			"expected breaker state Closed, got %v",
			state,
		)
	}
}

func TestErrorHandlerSelectedUpstreamOpensAfterFailures(
	t *testing.T,
) {
	cfg := breaker.Config{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		OpenTimeout:      time.Minute,
	}

	u := newTestUpstream(
		"test-upstream",
		cfg,
	)

	handler := newErrorHandler()

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		req = req.WithContext(
			IntoContext(
				req.Context(),
				u,
			),
		)

		rec := httptest.NewRecorder()

		handler(
			rec,
			req,
			errors.New("upstream connection failed"),
		)

		if rec.Code != http.StatusBadGateway {
			t.Fatalf(
				"request %d: expected status %d, got %d",
				i+1,
				http.StatusBadGateway,
				rec.Code,
			)
		}
	}

	if state := u.Breaker.State(); state != breaker.Open {
		t.Fatalf(
			"expected breaker state Open after failure threshold, got %v",
			state,
		)
	}

	if u.Breaker.Allow() {
		t.Fatal(
			"breaker should reject requests while Open",
		)
	}
}
