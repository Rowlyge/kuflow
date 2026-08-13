package breaker

import (
	"testing"
	"time"
)

func TestBreakerStartsClosed(t *testing.T) {
	breaker := New(Config{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenTimeout:      time.Second,
	})

	if state := breaker.State(); state != Closed {
		t.Fatalf(
			"expected initial state %v, got %v",
			Closed,
			state,
		)
	}
}

func TestBreakerAllowsRequestsWhenClosed(t *testing.T) {
	breaker := New(Config{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenTimeout:      time.Second,
	})

	if !breaker.Allow() {
		t.Fatal("closed breaker should allow requests")
	}
}

func TestBreakerOpensAfterFailureThreshold(t *testing.T) {
	breaker := New(Config{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenTimeout:      time.Second,
	})

	breaker.OnFailure()

	if state := breaker.State(); state != Closed {
		t.Fatalf(
			"expected state Closed after first failure, got %v",
			state,
		)
	}

	breaker.OnFailure()

	if state := breaker.State(); state != Closed {
		t.Fatalf(
			"expected state Closed after second failure, got %v",
			state,
		)
	}

	breaker.OnFailure()

	if state := breaker.State(); state != Open {
		t.Fatalf(
			"expected state Open after failure threshold, got %v",
			state,
		)
	}

	if breaker.Allow() {
		t.Fatal("open breaker should reject requests")
	}
}

func TestBreakerOpenTransitionsToHalfOpenAfterTimeout(t *testing.T) {
	breaker := New(Config{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		OpenTimeout:      10 * time.Millisecond,
	})

	breaker.OnFailure()

	if state := breaker.State(); state != Open {
		t.Fatalf(
			"expected state Open, got %v",
			state,
		)
	}

	time.Sleep(20 * time.Millisecond)

	if !breaker.Allow() {
		t.Fatal("breaker should allow a probe request after open timeout")
	}

	if state := breaker.State(); state != HalfOpen {
		t.Fatalf(
			"expected state HalfOpen, got %v",
			state,
		)
	}
}

func TestBreakerHalfOpenRequiresSuccessThreshold(t *testing.T) {
	breaker := New(Config{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		OpenTimeout:      0,
	})

	breaker.OnFailure()

	if !breaker.Allow() {
		t.Fatal("breaker should allow probe request")
	}

	if state := breaker.State(); state != HalfOpen {
		t.Fatalf(
			"expected state HalfOpen, got %v",
			state,
		)
	}

	breaker.OnSuccess()

	if state := breaker.State(); state != HalfOpen {
		t.Fatalf(
			"expected state HalfOpen after first success, got %v",
			state,
		)
	}

	breaker.OnSuccess()

	if state := breaker.State(); state != Closed {
		t.Fatalf(
			"expected state Closed after success threshold, got %v",
			state,
		)
	}

	if !breaker.Allow() {
		t.Fatal("closed breaker should allow requests")
	}
}

func TestBreakerHalfOpenFailureReopens(t *testing.T) {
	breaker := New(Config{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		OpenTimeout:      10 * time.Millisecond,
	})

	breaker.OnFailure()

	time.Sleep(20 * time.Millisecond)

	if !breaker.Allow() {
		t.Fatal("breaker should allow probe request")
	}

	if state := breaker.State(); state != HalfOpen {
		t.Fatalf(
			"expected state HalfOpen, got %v",
			state,
		)
	}

	breaker.OnFailure()

	if state := breaker.State(); state != Open {
		t.Fatalf(
			"expected state Open after half-open failure, got %v",
			state,
		)
	}

	if breaker.Allow() {
		t.Fatal("reopened breaker should reject requests")
	}
}

func TestBreakerSuccessResetsFailures(t *testing.T) {
	breaker := New(Config{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenTimeout:      time.Second,
	})

	breaker.OnFailure()
	breaker.OnFailure()

	if state := breaker.State(); state != Closed {
		t.Fatalf(
			"expected state Closed, got %v",
			state,
		)
	}

	breaker.OnSuccess()

	// Если счётчик ошибок не сбросился,
	// следующие две ошибки открыли бы Circuit.
	breaker.OnFailure()

	if state := breaker.State(); state != Closed {
		t.Fatalf(
			"expected state Closed after reset, got %v",
			state,
		)
	}

	breaker.OnFailure()

	if state := breaker.State(); state != Closed {
		t.Fatalf(
			"expected state Closed after second post-reset failure, got %v",
			state,
		)
	}

	breaker.OnFailure()

	if state := breaker.State(); state != Open {
		t.Fatalf(
			"expected state Open after three post-reset failures, got %v",
			state,
		)
	}
}

func TestBreakerDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.FailureThreshold != 5 {
		t.Fatalf(
			"expected failure threshold %d, got %d",
			5,
			cfg.FailureThreshold,
		)
	}

	if cfg.SuccessThreshold != 2 {
		t.Fatalf(
			"expected success threshold %d, got %d",
			2,
			cfg.SuccessThreshold,
		)
	}

	if cfg.OpenTimeout != 30*time.Second {
		t.Fatalf(
			"expected open timeout %v, got %v",
			30*time.Second,
			cfg.OpenTimeout,
		)
	}
}
