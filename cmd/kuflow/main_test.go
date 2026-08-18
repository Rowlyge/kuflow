package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type testShutdownServer struct {
	mu sync.Mutex

	shutdownCalled bool
	shutdownCtx    context.Context

	shutdownErr error

	events *[]string
}

func (s *testShutdownServer) Shutdown(
	ctx context.Context,
) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.shutdownCalled = true
	s.shutdownCtx = ctx

	if s.events != nil {
		*s.events = append(
			*s.events,
			"server.shutdown",
		)
	}

	return s.shutdownErr
}

type testShutdownApp struct {
	mu sync.Mutex

	closeCalled bool
	closeCtx    context.Context

	closeErr error

	events *[]string
	server *testShutdownServer
}

func (a *testShutdownApp) Close(
	ctx context.Context,
) error {

	a.mu.Lock()
	defer a.mu.Unlock()

	a.closeCalled = true
	a.closeCtx = ctx

	if a.events != nil {
		*a.events = append(
			*a.events,
			"app.close",
		)
	}

	// Если server ещё не завершил Shutdown(),
	// порядок shutdown нарушен.
	if a.server != nil {
		a.server.mu.Lock()
		serverShutdownCalled := a.server.shutdownCalled
		a.server.mu.Unlock()

		if !serverShutdownCalled {
			return errors.New(
				"application closed before HTTP server",
			)
		}
	}

	return a.closeErr
}

func TestShutdownStopsServerBeforeApplication(t *testing.T) {
	events := make([]string, 0, 2)

	httpServer := &testShutdownServer{
		events: &events,
	}

	application := &testShutdownApp{
		events: &events,
		server: httpServer,
	}

	err := shutdown(
		httpServer,
		application,
	)
	if err != nil {
		t.Fatalf(
			"shutdown() returned error: %v",
			err,
		)
	}

	if got := httpServer.shutdownCalled; !got {
		t.Fatal(
			"HTTP server Shutdown() was not called",
		)
	}

	if got := application.closeCalled; !got {
		t.Fatal(
			"Application Close() was not called",
		)
	}

	want := []string{
		"server.shutdown",
		"app.close",
	}

	if len(events) != len(want) {
		t.Fatalf(
			"events = %v, want %v",
			events,
			want,
		)
	}

	for i := range want {
		if events[i] != want[i] {
			t.Fatalf(
				"events = %v, want %v",
				events,
				want,
			)
		}
	}
}

func TestShutdownUsesIndependentContexts(t *testing.T) {
	httpServer := &testShutdownServer{}

	application := &testShutdownApp{}

	err := shutdown(
		httpServer,
		application,
	)
	if err != nil {
		t.Fatalf(
			"shutdown() returned error: %v",
			err,
		)
	}

	if httpServer.shutdownCtx == nil {
		t.Fatal(
			"HTTP server Shutdown() received nil context",
		)
	}

	if application.closeCtx == nil {
		t.Fatal(
			"Application Close() received nil context",
		)
	}

	if httpServer.shutdownCtx == application.closeCtx {
		t.Fatal(
			"HTTP server and Application must use independent contexts",
		)
	}

	if deadline, ok := httpServer.shutdownCtx.Deadline(); !ok {
		t.Fatal(
			"HTTP server shutdown context has no deadline",
		)
	} else if time.Until(deadline) <= 0 {
		t.Fatal(
			"HTTP server shutdown context deadline already expired",
		)
	}

	if deadline, ok := application.closeCtx.Deadline(); !ok {
		t.Fatal(
			"Application shutdown context has no deadline",
		)
	} else if time.Until(deadline) <= 0 {
		t.Fatal(
			"Application shutdown context deadline already expired",
		)
	}
}

func TestShutdownContinuesWhenServerShutdownFails(
	t *testing.T,
) {
	serverErr := errors.New("server shutdown failed")

	httpServer := &testShutdownServer{
		shutdownErr: serverErr,
	}

	application := &testShutdownApp{}

	err := shutdown(
		httpServer,
		application,
	)

	if !errors.Is(err, serverErr) {
		t.Fatalf(
			"shutdown() error = %v, want %v",
			err,
			serverErr,
		)
	}

	if !application.closeCalled {
		t.Fatal(
			"Application Close() was not called after server shutdown failure",
		)
	}
}

func TestShutdownReturnsApplicationError(
	t *testing.T,
) {
	appErr := errors.New("application shutdown failed")

	httpServer := &testShutdownServer{}

	application := &testShutdownApp{
		closeErr: appErr,
	}

	err := shutdown(
		httpServer,
		application,
	)

	if !errors.Is(err, appErr) {
		t.Fatalf(
			"shutdown() error = %v, want %v",
			err,
			appErr,
		)
	}
}

func TestShutdownJoinsServerAndApplicationErrors(
	t *testing.T,
) {
	serverErr := errors.New("server shutdown failed")
	appErr := errors.New("application shutdown failed")

	httpServer := &testShutdownServer{
		shutdownErr: serverErr,
	}

	application := &testShutdownApp{
		closeErr: appErr,
	}

	err := shutdown(
		httpServer,
		application,
	)

	if !errors.Is(err, serverErr) {
		t.Fatalf(
			"shutdown() error does not contain server error: %v",
			err,
		)
	}

	if !errors.Is(err, appErr) {
		t.Fatalf(
			"shutdown() error does not contain application error: %v",
			err,
		)
	}
}

func TestShutdownDoesNotWaitAfterExpiredServerContext(
	t *testing.T,
) {
	serverErr := context.DeadlineExceeded

	httpServer := &testShutdownServer{
		shutdownErr: serverErr,
	}

	application := &testShutdownApp{}

	start := time.Now()

	err := shutdown(
		httpServer,
		application,
	)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal(
			"shutdown() returned nil, want server shutdown error",
		)
	}

	if !errors.Is(err, serverErr) {
		t.Fatalf(
			"shutdown() error = %v, want context deadline exceeded",
			err,
		)
	}

	if !application.closeCalled {
		t.Fatal(
			"Application Close() was not called after server timeout",
		)
	}

	if elapsed > time.Second {
		t.Fatalf(
			"shutdown() took too long: %v",
			elapsed,
		)
	}
}
