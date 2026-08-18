package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Rowlyge/kuflow/internal/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := config.ServerConfig{
		Port:         "0",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}

	handler := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("ok"))
	})

	return newWithHandler(handler, cfg)
}

func waitForServer(
	t *testing.T,
	s *Server,
) string {
	t.Helper()

	deadline := time.Now().Add(time.Second)

	for time.Now().Before(deadline) {
		addr := s.Addr()

		if addr != nil {
			return addr.String()
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("HTTP server listener was not created")

	return ""
}

func TestServer_StartAcceptsRequests(t *testing.T) {
	s := newTestServer(t)

	startErr := make(chan error, 1)

	go func() {
		startErr <- s.Start()
	}()

	addr := waitForServer(t, s)

	url := "http://" + addr

	var resp *http.Response
	var err error

	deadline := time.Now().Add(time.Second)

	for time.Now().Before(deadline) {
		resp, err = http.Get(url)

		if err == nil {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf(
			"failed to send HTTP request: %v",
			err,
		)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(
			"failed to read response body: %v",
			err,
		)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	if got := string(body); got != "ok" {
		t.Fatalf(
			"response body = %q, want %q",
			got,
			"ok",
		)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		t.Fatalf(
			"Shutdown() failed: %v",
			err,
		)
	}

	select {
	case err := <-startErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf(
				"Start() returned %v, want %v",
				err,
				http.ErrServerClosed,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("Start() did not return after Shutdown()")
	}
}

func TestServer_StartReturnsServerClosedAfterShutdown(
	t *testing.T,
) {
	s := newTestServer(t)

	startErr := make(chan error, 1)

	go func() {
		startErr <- s.Start()
	}()

	waitForServer(t, s)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf(
			"Shutdown() failed: %v",
			err,
		)
	}

	select {
	case err := <-startErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf(
				"Start() returned %v, want %v",
				err,
				http.ErrServerClosed,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("Start() did not return after Shutdown()")
	}
}

func TestServer_ShutdownStopsServer(t *testing.T) {
	s := newTestServer(t)

	startErr := make(chan error, 1)

	go func() {
		startErr <- s.Start()
	}()

	addr := waitForServer(t, s)

	url := "http://" + addr

	// Убеждаемся, что сервер действительно принимает запросы.
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf(
			"server is not accepting requests: %v",
			err,
		)
	}
	resp.Body.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf(
			"Shutdown() failed: %v",
			err,
		)
	}

	select {
	case err := <-startErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf(
				"Start() returned %v, want %v",
				err,
				http.ErrServerClosed,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("server did not stop after Shutdown()")
	}

	// После Shutdown новые подключения должны быть невозможны.
	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}

	_, err = client.Get(url)
	if err == nil {
		t.Fatal("request succeeded after server shutdown")
	}
}

func TestServer_ShutdownWithExpiredContext(
	t *testing.T,
) {
	s := newTestServer(t)

	startErr := make(chan error, 1)

	go func() {
		startErr <- s.Start()
	}()

	waitForServer(t, s)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		0,
	)
	defer cancel()

	start := time.Now()

	err := s.Shutdown(ctx)

	elapsed := time.Since(start)

	if elapsed > time.Second {
		t.Fatalf(
			"Shutdown() took too long with expired context: %v",
			elapsed,
		)
	}

	// Для уже остановленного/быстро останавливающегося HTTP-сервера
	// Shutdown может вернуть как ошибку context, так и nil.
	if err != nil &&
		!errors.Is(err, context.DeadlineExceeded) &&
		!errors.Is(err, context.Canceled) {

		t.Fatalf(
			"Shutdown() returned unexpected error: %v",
			err,
		)
	}

	select {
	case err := <-startErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf(
				"Start() returned %v, want %v",
				err,
				http.ErrServerClosed,
			)
		}

	case <-time.After(time.Second):
		t.Fatal(
			"Start() did not return after Shutdown()",
		)
	}
}

func TestServer_StartFailsWithInvalidAddress(
	t *testing.T,
) {
	cfg := config.ServerConfig{
		Port:         "invalid-port",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}

	handler := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	s := newWithHandler(handler, cfg)

	err := s.Start()

	if err == nil {
		t.Fatal(
			"Start() returned nil, want listen error",
		)
	}

	if strings.Contains(
		err.Error(),
		"invalid-port",
	) == false {
		t.Fatalf(
			"Start() returned unexpected error: %v",
			err,
		)
	}
}
