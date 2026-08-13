package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Rowlyge/kuflow/internal/breaker"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

func TestModifyResponseRecordsSuccess(t *testing.T) {
	cfg := breaker.Config{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      10 * time.Millisecond,
	}

	upstream := newTestUpstream(
		"test-upstream",
		cfg,
	)

	// Переводим именно upstream.Breaker в Open.
	upstream.Breaker.OnFailure()

	if state := upstream.Breaker.State(); state != breaker.Open {
		t.Fatalf(
			"expected breaker state %v after failure, got %v",
			breaker.Open,
			state,
		)
	}

	// Ждём перехода Open -> HalfOpen.
	time.Sleep(
		cfg.OpenTimeout + time.Millisecond,
	)

	if !upstream.Breaker.Allow() {
		t.Fatal(
			"expected half-open request to be allowed",
		)
	}

	if state := upstream.Breaker.State(); state != breaker.HalfOpen {
		t.Fatalf(
			"expected breaker state %v, got %v",
			breaker.HalfOpen,
			state,
		)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req = req.WithContext(
		IntoContext(
			req.Context(),
			upstream,
		),
	)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Request:    req,
	}

	if err := newModifyResponse()(resp); err != nil {
		t.Fatalf(
			"modify response returned error: %v",
			err,
		)
	}

	if state := upstream.Breaker.State(); state != breaker.Closed {
		t.Fatalf(
			"expected breaker state %v after successful response, got %v",
			breaker.Closed,
			state,
		)
	}
}

func TestModifyResponseRecordsFailureForServerError(t *testing.T) {
	cfg := breaker.Config{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      time.Minute,
	}

	upstream := newTestUpstream(
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
			upstream,
		),
	)

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Header:     make(http.Header),
		Request:    req,
	}

	if err := newModifyResponse()(resp); err != nil {
		t.Fatalf(
			"modify response returned error: %v",
			err,
		)
	}

	if state := upstream.Breaker.State(); state != breaker.Open {
		t.Fatalf(
			"expected breaker state %v after 5xx response, got %v",
			breaker.Open,
			state,
		)
	}
}

func TestModifyResponseRecordsSuccessForClientError(t *testing.T) {
	cfg := breaker.Config{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		OpenTimeout:      time.Minute,
	}

	upstream := newTestUpstream(
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
			upstream,
		),
	)

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     make(http.Header),
		Request:    req,
	}

	if err := newModifyResponse()(resp); err != nil {
		t.Fatalf(
			"modify response returned error: %v",
			err,
		)
	}

	if state := upstream.Breaker.State(); state != breaker.Closed {
		t.Fatalf(
			"expected breaker state %v after 4xx response, got %v",
			breaker.Closed,
			state,
		)
	}
}

func TestModifyResponseAddsKuFlowHeader(t *testing.T) {
	b := breaker.New(breaker.DefaultConfig())

	u := &upstream.Upstream{
		Name:    "test-upstream",
		Breaker: b,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"http://upstream.test/",
		nil,
	)

	req = req.WithContext(
		IntoContext(
			req.Context(),
			u,
		),
	)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Request:    req,
	}

	modifyResponse := newModifyResponse()

	if err := modifyResponse(resp); err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if got := resp.Header.Get("X-KuFlow"); got != "true" {
		t.Fatalf(
			"expected X-KuFlow header to be %q, got %q",
			"true",
			got,
		)
	}
}
