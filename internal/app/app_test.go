package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type testLifecycleWorker struct {
	mu sync.Mutex

	started int
	waited  int

	startedCh     chan struct{}
	waitStartedCh chan struct{}
	stoppedCh     chan struct{}

	startedOnce     sync.Once
	waitStartedOnce sync.Once
	stoppedOnce     sync.Once
}

func newTestLifecycleWorker() *testLifecycleWorker {
	return &testLifecycleWorker{
		startedCh:     make(chan struct{}),
		waitStartedCh: make(chan struct{}),
		stoppedCh:     make(chan struct{}),
	}
}

func (w *testLifecycleWorker) Start(ctx context.Context) {
	w.mu.Lock()

	w.started++

	w.startedOnce.Do(func() {
		close(w.startedCh)
	})

	w.mu.Unlock()

	// Worker exits its Start phase after lifecycle cancellation.
	<-ctx.Done()

	w.mu.Lock()
	w.waited++
	w.mu.Unlock()
}

func (w *testLifecycleWorker) Wait() {
	w.waitStartedOnce.Do(func() {
		close(w.waitStartedCh)
	})

	// Wait intentionally blocks until Stop() is called.
	//
	// This allows tests to distinguish:
	// - worker has stopped its Start phase;
	// - worker lifecycle is fully finished.
	<-w.stoppedCh
}

func (w *testLifecycleWorker) Stop() {
	w.stoppedOnce.Do(func() {
		close(w.stoppedCh)
	})
}

func (w *testLifecycleWorker) Started() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.started
}

func (w *testLifecycleWorker) Waited() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.waited
}

func waitFor(
	t *testing.T,
	ch <-chan struct{},
) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lifecycle event")
	}
}

func TestApp_StartStartsAllWorkers(t *testing.T) {
	worker1 := newTestLifecycleWorker()
	worker2 := newTestLifecycleWorker()
	worker3 := newTestLifecycleWorker()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	app := &App{
		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,

		lifecycleWorkers: []lifecycleWorker{
			worker1,
			worker2,
			worker3,
		},
	}

	app.Start()

	waitFor(t, worker1.startedCh)
	waitFor(t, worker2.startedCh)
	waitFor(t, worker3.startedCh)

	if got := worker1.Started(); got != 1 {
		t.Fatalf("worker1 started %d times, want 1", got)
	}

	if got := worker2.Started(); got != 1 {
		t.Fatalf("worker2 started %d times, want 1", got)
	}

	if got := worker3.Started(); got != 1 {
		t.Fatalf("worker3 started %d times, want 1", got)
	}

	cancel()

	// Start() has returned for all workers, but Wait() is still blocked.
	// Release Wait() explicitly so the test can finish.
	worker1.Stop()
	worker2.Stop()
	worker3.Stop()

	app.lifecycleWG.Wait()

	if got := worker1.Waited(); got != 1 {
		t.Fatalf("worker1 waited %d times, want 1", got)
	}

	if got := worker2.Waited(); got != 1 {
		t.Fatalf("worker2 waited %d times, want 1", got)
	}

	if got := worker3.Waited(); got != 1 {
		t.Fatalf("worker3 waited %d times, want 1", got)
	}
}

func TestApp_StartIsIdempotent(t *testing.T) {
	worker := newTestLifecycleWorker()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	app := &App{
		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,

		lifecycleWorkers: []lifecycleWorker{
			worker,
		},
	}

	app.Start()
	app.Start()
	app.Start()

	waitFor(t, worker.startedCh)

	if got := worker.Started(); got != 1 {
		t.Fatalf(
			"worker started %d times, want exactly 1",
			got,
		)
	}

	cancel()

	// Release Wait().
	worker.Stop()

	app.lifecycleWG.Wait()

	if got := worker.Waited(); got != 1 {
		t.Fatalf(
			"worker waited %d times, want exactly 1",
			got,
		)
	}
}

func TestApp_CloseStopsWorkers(t *testing.T) {
	worker1 := newTestLifecycleWorker()
	worker2 := newTestLifecycleWorker()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	app := &App{
		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,

		lifecycleWorkers: []lifecycleWorker{
			worker1,
			worker2,
		},
	}

	app.Start()

	waitFor(t, worker1.startedCh)
	waitFor(t, worker2.startedCh)

	// Release Wait() after Close cancels the lifecycle.
	go func() {
		waitFor(t, worker1.waitStartedCh)
		worker1.Stop()
	}()

	go func() {
		waitFor(t, worker2.waitStartedCh)
		worker2.Stop()
	}()

	done := make(chan error, 1)

	go func() {
		done <- app.Close(context.Background())
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("App.Close() returned error: %v", err)
		}

	case <-time.After(time.Second):
		t.Fatal("App.Close() did not return")
	}

	if got := worker1.Waited(); got != 1 {
		t.Fatalf("worker1 waited %d times, want 1", got)
	}

	if got := worker2.Waited(); got != 1 {
		t.Fatalf("worker2 waited %d times, want 1", got)
	}
}

func TestApp_CloseIsIdempotent(t *testing.T) {
	worker := newTestLifecycleWorker()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	app := &App{
		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,

		lifecycleWorkers: []lifecycleWorker{
			worker,
		},
	}

	app.Start()

	waitFor(t, worker.startedCh)

	// Release Wait() when it starts.
	go func() {
		waitFor(t, worker.waitStartedCh)
		worker.Stop()
	}()

	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("first Close() failed: %v", err)
	}

	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("second Close() failed: %v", err)
	}

	if got := worker.Started(); got != 1 {
		t.Fatalf("worker started %d times, want 1", got)
	}

	if got := worker.Waited(); got != 1 {
		t.Fatalf("worker waited %d times, want 1", got)
	}
}

func TestApp_CloseRespectsContextTimeout(t *testing.T) {
	worker := newTestLifecycleWorker()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	app := &App{
		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,

		lifecycleWorkers: []lifecycleWorker{
			worker,
		},
	}

	app.Start()

	waitFor(t, worker.startedCh)

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer shutdownCancel()

	start := time.Now()

	err := app.Close(shutdownCtx)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("App.Close() returned nil, want timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"App.Close() returned error %v, want context deadline exceeded",
			err,
		)
	}

	if elapsed < 40*time.Millisecond {
		t.Fatalf(
			"App.Close() returned too early: %v",
			elapsed,
		)
	}

	if got := worker.Waited(); got != 1 {
		t.Fatalf(
			"worker Start() completed %d times, want 1",
			got,
		)
	}

	// Intentionally leave Wait() blocked.
	// This proves that Close() returned because of its
	// shutdown context deadline.
}

func TestApp_CloseWaitsForWorkersWhenTheyStop(t *testing.T) {
	worker := newTestLifecycleWorker()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	app := &App{
		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,

		lifecycleWorkers: []lifecycleWorker{
			worker,
		},
	}

	app.Start()

	waitFor(t, worker.startedCh)

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer shutdownCancel()

	done := make(chan error, 1)

	go func() {
		done <- app.Close(shutdownCtx)
	}()

	// Close() отменяет lifecycle context.
	// После этого Start() завершается и worker входит в Wait().
	waitFor(t, worker.waitStartedCh)

	// Пока Wait() заблокирован, Close() не должен вернуться.
	select {
	case err := <-done:
		t.Fatalf(
			"App.Close() returned before worker stopped: %v",
			err,
		)

	case <-time.After(20 * time.Millisecond):
	}

	// Теперь разрешаем worker завершить Wait().
	worker.Stop()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf(
				"App.Close() returned error: %v",
				err,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("App.Close() did not wait for worker to stop")
	}

	if got := worker.Waited(); got != 1 {
		t.Fatalf(
			"worker Start() completed %d times, want 1",
			got,
		)
	}
}
