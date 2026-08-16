package app

import (
	"context"
	"sync"
	"testing"
	"time"
)

type testLifecycleWorker struct {
	mu sync.Mutex

	started int
	waited  int

	startedCh chan struct{}
	stoppedCh chan struct{}

	once sync.Once
}

func newTestLifecycleWorker() *testLifecycleWorker {
	return &testLifecycleWorker{
		startedCh: make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
}

func (w *testLifecycleWorker) Start(ctx context.Context) {
	w.mu.Lock()

	w.started++

	w.once.Do(func() {
		close(w.startedCh)
	})

	w.mu.Unlock()

	<-ctx.Done()

	w.mu.Lock()
	w.waited++
	close(w.stoppedCh)
	w.mu.Unlock()
}

func (w *testLifecycleWorker) Wait() {
	<-w.stoppedCh
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
		lifecycleCtx: ctx,

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
		lifecycleCtx: ctx,

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

	done := make(chan struct{})

	go func() {
		_ = app.Close()
		close(done)
	}()

	select {
	case <-done:
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

	if err := app.Close(); err != nil {
		t.Fatalf("first Close() failed: %v", err)
	}

	if err := app.Close(); err != nil {
		t.Fatalf("second Close() failed: %v", err)
	}

	if got := worker.Started(); got != 1 {
		t.Fatalf("worker started %d times, want 1", got)
	}

	if got := worker.Waited(); got != 1 {
		t.Fatalf("worker waited %d times, want 1", got)
	}
}
