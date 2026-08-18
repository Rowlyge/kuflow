package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

const testShutdownTimeout = 50 * time.Millisecond

type testShutdownServer struct {
	mu sync.Mutex

	shutdownCalled       bool
	shutdownCtx          context.Context
	shutdownCtxErrAtCall error

	shutdownErr   error
	shutdownBlock bool

	events *[]string
}

func (s *testShutdownServer) Shutdown(
	ctx context.Context,
) error {
	s.mu.Lock()

	s.shutdownCalled = true
	s.shutdownCtx = ctx
	s.shutdownCtxErrAtCall = ctx.Err()

	if s.events != nil {
		*s.events = append(
			*s.events,
			"server.shutdown",
		)
	}

	block := s.shutdownBlock
	err := s.shutdownErr

	s.mu.Unlock()

	if block {
		<-ctx.Done()
		return ctx.Err()
	}

	return err
}

func (s *testShutdownServer) isShutdownCalled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.shutdownCalled
}

func (s *testShutdownServer) getShutdownContext() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.shutdownCtx
}

func (s *testShutdownServer) getShutdownContextErrAtCall() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.shutdownCtxErrAtCall
}

type testShutdownApp struct {
	mu sync.Mutex

	closeCalled       bool
	closeCtx          context.Context
	closeCtxErrAtCall error

	closeErr   error
	closeBlock bool

	events *[]string
	server *testShutdownServer
}

func (a *testShutdownApp) Close(
	ctx context.Context,
) error {
	a.mu.Lock()

	a.closeCalled = true
	a.closeCtx = ctx
	a.closeCtxErrAtCall = ctx.Err()

	if a.events != nil {
		*a.events = append(
			*a.events,
			"app.close",
		)
	}

	server := a.server
	block := a.closeBlock
	err := a.closeErr

	a.mu.Unlock()

	if server != nil {
		if !server.isShutdownCalled() {
			return errors.New(
				"application closed before HTTP server",
			)
		}
	}

	if block {
		<-ctx.Done()
		return ctx.Err()
	}

	return err
}

func (a *testShutdownApp) isCloseCalled() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.closeCalled
}

func (a *testShutdownApp) getCloseContext() context.Context {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.closeCtx
}

func (a *testShutdownApp) getCloseContextErrAtCall() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.closeCtxErrAtCall
}

func runShutdown(
	httpServer *testShutdownServer,
	application *testShutdownApp,
	serverTimeout time.Duration,
	appTimeout time.Duration,
) error {
	return shutdown(
		httpServer,
		application,
		serverTimeout,
		appTimeout,
	)
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

	err := runShutdown(
		httpServer,
		application,
		testShutdownTimeout,
		testShutdownTimeout,
	)

	if err != nil {
		t.Fatalf(
			"shutdown() returned error: %v",
			err,
		)
	}

	if !httpServer.isShutdownCalled() {
		t.Fatal(
			"HTTP server Shutdown() was not called",
		)
	}

	if !application.isCloseCalled() {
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

func TestShutdownWaitsForServerBeforeApplication(t *testing.T) {
	events := make([]string, 0, 2)

	httpServer := &testShutdownServer{
		events:        &events,
		shutdownBlock: true,
	}

	application := &testShutdownApp{
		events: &events,
		server: httpServer,
	}

	done := make(chan error, 1)

	go func() {
		done <- runShutdown(
			httpServer,
			application,
			testShutdownTimeout,
			testShutdownTimeout,
		)
	}()

	select {
	case <-done:
		t.Fatal(
			"shutdown() returned before HTTP server shutdown completed",
		)

	case <-time.After(10 * time.Millisecond):
	}

	if application.isCloseCalled() {
		t.Fatal(
			"Application Close() was called before HTTP server shutdown completed",
		)
	}

	select {
	case err := <-done:
		if !errors.Is(
			err,
			context.DeadlineExceeded,
		) {
			t.Fatalf(
				"shutdown() error = %v, want context deadline exceeded",
				err,
			)
		}

	case <-time.After(500 * time.Millisecond):
		t.Fatal(
			"shutdown() did not finish after server timeout",
		)
	}

	if !application.isCloseCalled() {
		t.Fatal(
			"Application Close() was not called after server timeout",
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

	err := runShutdown(
		httpServer,
		application,
		testShutdownTimeout,
		testShutdownTimeout,
	)

	if err != nil {
		t.Fatalf(
			"shutdown() returned error: %v",
			err,
		)
	}

	serverCtx := httpServer.getShutdownContext()
	appCtx := application.getCloseContext()

	if httpServer.getShutdownContextErrAtCall() != nil {
		t.Fatalf(
			"HTTP server shutdown context was already cancelled when Shutdown() was called: %v",
			httpServer.getShutdownContextErrAtCall(),
		)
	}

	if application.getCloseContextErrAtCall() != nil {
		t.Fatalf(
			"Application shutdown context was already cancelled when Close() was called: %v",
			application.getCloseContextErrAtCall(),
		)
	}

	if serverCtx == nil {
		t.Fatal(
			"HTTP server Shutdown() received nil context",
		)
	}

	if appCtx == nil {
		t.Fatal(
			"Application Close() received nil context",
		)
	}

	if serverCtx == appCtx {
		t.Fatal(
			"HTTP server and Application must use independent contexts",
		)
	}

	serverDeadline, serverOK := serverCtx.Deadline()

	if !serverOK {
		t.Fatal(
			"HTTP server shutdown context has no deadline",
		)
	}

	if serverDeadline.Before(time.Now()) {
		t.Fatal(
			"HTTP server shutdown context deadline already expired",
		)
	}

	appDeadline, appOK := appCtx.Deadline()

	if !appOK {
		t.Fatal(
			"Application shutdown context has no deadline",
		)
	}

	if appDeadline.Before(time.Now()) {
		t.Fatal(
			"Application shutdown context deadline already expired",
		)
	}
}

func TestShutdownContinuesWhenServerShutdownFails(t *testing.T) {
	serverErr := errors.New(
		"server shutdown failed",
	)

	httpServer := &testShutdownServer{
		shutdownErr: serverErr,
	}

	application := &testShutdownApp{}

	err := runShutdown(
		httpServer,
		application,
		testShutdownTimeout,
		testShutdownTimeout,
	)

	if !errors.Is(err, serverErr) {
		t.Fatalf(
			"shutdown() error = %v, want %v",
			err,
			serverErr,
		)
	}

	if !application.isCloseCalled() {
		t.Fatal(
			"Application Close() was not called after server shutdown failure",
		)
	}
}

func TestShutdownReturnsApplicationError(t *testing.T) {
	appErr := errors.New(
		"application shutdown failed",
	)

	httpServer := &testShutdownServer{}

	application := &testShutdownApp{
		closeErr: appErr,
	}

	err := runShutdown(
		httpServer,
		application,
		testShutdownTimeout,
		testShutdownTimeout,
	)

	if !errors.Is(err, appErr) {
		t.Fatalf(
			"shutdown() error = %v, want %v",
			err,
			appErr,
		)
	}
}

func TestShutdownJoinsServerAndApplicationErrors(t *testing.T) {
	serverErr := errors.New(
		"server shutdown failed",
	)

	appErr := errors.New(
		"application shutdown failed",
	)

	httpServer := &testShutdownServer{
		shutdownErr: serverErr,
	}

	application := &testShutdownApp{
		closeErr: appErr,
	}

	err := runShutdown(
		httpServer,
		application,
		testShutdownTimeout,
		testShutdownTimeout,
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

func TestShutdownStopsApplicationAfterServerContextExpires(t *testing.T) {
	httpServer := &testShutdownServer{
		shutdownBlock: true,
	}

	application := &testShutdownApp{}

	start := time.Now()

	err := runShutdown(
		httpServer,
		application,
		testShutdownTimeout,
		testShutdownTimeout,
	)

	elapsed := time.Since(start)

	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"shutdown() error = %v, want context deadline exceeded",
			err,
		)
	}

	if !application.isCloseCalled() {
		t.Fatal(
			"Application Close() was not called after server context expired",
		)
	}

	if application.getCloseContextErrAtCall() != nil {
		t.Fatalf(
			"Application shutdown context was already cancelled: %v",
			application.getCloseContextErrAtCall(),
		)
	}

	if elapsed < testShutdownTimeout {
		t.Fatalf(
			"shutdown() finished too early: %v",
			elapsed,
		)
	}

	if elapsed > 500*time.Millisecond {
		t.Fatalf(
			"shutdown() took too long: %v",
			elapsed,
		)
	}
}

func TestShutdownStopsApplicationAfterServerError(t *testing.T) {
	serverErr := errors.New(
		"server shutdown failed",
	)

	httpServer := &testShutdownServer{
		shutdownErr: serverErr,
	}

	application := &testShutdownApp{}

	err := runShutdown(
		httpServer,
		application,
		testShutdownTimeout,
		testShutdownTimeout,
	)

	if !errors.Is(err, serverErr) {
		t.Fatalf(
			"shutdown() error = %v, want %v",
			err,
			serverErr,
		)
	}

	if !application.isCloseCalled() {
		t.Fatal(
			"Application Close() was not called after server error",
		)
	}
}

func TestShutdownApplicationContextIsIndependentFromServerTimeout(
	t *testing.T,
) {
	httpServer := &testShutdownServer{
		shutdownBlock: true,
	}

	application := &testShutdownApp{}

	start := time.Now()

	err := runShutdown(
		httpServer,
		application,
		30*time.Millisecond,
		100*time.Millisecond,
	)

	elapsed := time.Since(start)

	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"shutdown() error = %v, want context deadline exceeded",
			err,
		)
	}

	if !application.isCloseCalled() {
		t.Fatal(
			"Application Close() was not called",
		)
	}

	if application.getCloseContextErrAtCall() != nil {
		t.Fatalf(
			"Application shutdown context was already cancelled: %v",
			application.getCloseContextErrAtCall(),
		)
	}

	if elapsed < 30*time.Millisecond {
		t.Fatalf(
			"shutdown() finished too early: %v",
			elapsed,
		)
	}

	if elapsed > 500*time.Millisecond {
		t.Fatalf(
			"shutdown() took too long: %v",
			elapsed,
		)
	}
}

func TestShutdownApplicationTimeoutIsIndependent(t *testing.T) {
	httpServer := &testShutdownServer{}

	application := &testShutdownApp{
		closeBlock: true,
	}

	start := time.Now()

	err := runShutdown(
		httpServer,
		application,
		100*time.Millisecond,
		30*time.Millisecond,
	)

	elapsed := time.Since(start)

	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"shutdown() error = %v, want context deadline exceeded",
			err,
		)
	}

	if elapsed < 30*time.Millisecond {
		t.Fatalf(
			"shutdown() finished too early: %v",
			elapsed,
		)
	}

	if elapsed > 500*time.Millisecond {
		t.Fatalf(
			"shutdown() took too long: %v",
			elapsed,
		)
	}

	if application.getCloseContextErrAtCall() != nil {
		t.Fatalf(
			"Application shutdown context was already cancelled: %v",
			application.getCloseContextErrAtCall(),
		)
	}
}
