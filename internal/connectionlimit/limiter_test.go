package connectionlimit

import (
	"sync"
	"testing"
)

func TestLimiterAcquire(t *testing.T) {
	limiter := New(2)

	if !limiter.Acquire("key-1") {
		t.Fatal("first Acquire() should succeed")
	}

	if !limiter.Acquire("key-1") {
		t.Fatal("second Acquire() should succeed")
	}

	if limiter.Acquire("key-1") {
		t.Fatal("third Acquire() should fail after reaching limit")
	}

	if got := limiter.Active("key-1"); got != 2 {
		t.Fatalf("expected 2 active connections, got %d", got)
	}
}

func TestLimiterRelease(t *testing.T) {
	limiter := New(2)

	limiter.Acquire("key-1")
	limiter.Acquire("key-1")

	limiter.Release("key-1")

	if got := limiter.Active("key-1"); got != 1 {
		t.Fatalf("expected 1 active connection after Release(), got %d", got)
	}

	if !limiter.Acquire("key-1") {
		t.Fatal("Acquire() should succeed after releasing a slot")
	}

	if got := limiter.Active("key-1"); got != 2 {
		t.Fatalf("expected 2 active connections, got %d", got)
	}
}

func TestLimiterDifferentAPIKeys(t *testing.T) {
	limiter := New(1)

	if !limiter.Acquire("key-1") {
		t.Fatal("key-1 should acquire a connection")
	}

	if !limiter.Acquire("key-2") {
		t.Fatal("key-2 should acquire a connection independently")
	}

	if limiter.Acquire("key-1") {
		t.Fatal("key-1 should be limited")
	}

	if limiter.Acquire("key-2") {
		t.Fatal("key-2 should be limited")
	}
}

func TestLimiterReleaseRemovesEmptyEntry(t *testing.T) {
	limiter := New(1)

	if !limiter.Acquire("key-1") {
		t.Fatal("Acquire() should succeed")
	}

	limiter.Release("key-1")

	if got := limiter.Active("key-1"); got != 0 {
		t.Fatalf("expected 0 active connections, got %d", got)
	}

	if len(limiter.active) != 0 {
		t.Fatalf("expected empty active map, got %d entries", len(limiter.active))
	}
}

func TestLimiterReleaseDoesNotGoNegative(t *testing.T) {
	limiter := New(1)

	limiter.Release("key-1")

	if got := limiter.Active("key-1"); got != 0 {
		t.Fatalf("expected 0 active connections, got %d", got)
	}

	limiter.Release("key-1")

	if got := limiter.Active("key-1"); got != 0 {
		t.Fatalf("expected 0 active connections after second Release(), got %d", got)
	}
}

func TestLimiterDefaultLimit(t *testing.T) {
	limiter := New(0)

	if !limiter.Acquire("key-1") {
		t.Fatal("Acquire() should succeed with normalized default limit")
	}

	if limiter.Acquire("key-1") {
		t.Fatal("second Acquire() should fail with limit 1")
	}
}

func TestLimiterConcurrentAccess(t *testing.T) {
	const (
		maxConnections = 10
		goroutines     = 100
	)

	limiter := New(maxConnections)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		active  int
		maxSeen int
	)

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			if !limiter.Acquire("key-1") {
				return
			}

			mu.Lock()

			active++

			if active > maxSeen {
				maxSeen = active
			}

			mu.Unlock()

			limiter.Release("key-1")

			mu.Lock()
			active--
			mu.Unlock()
		}()
	}

	wg.Wait()

	if maxSeen > maxConnections {
		t.Fatalf(
			"connection limit exceeded: max seen %d, limit %d",
			maxSeen,
			maxConnections,
		)
	}

	if got := limiter.Active("key-1"); got != 0 {
		t.Fatalf("expected 0 active connections after all goroutines, got %d", got)
	}
}
