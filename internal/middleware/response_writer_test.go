package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResponseWriter_WriteHeaderOnlyFirstStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseWriter(rec)

	w.WriteHeader(http.StatusCreated)
	w.WriteHeader(http.StatusInternalServerError)

	if got := w.StatusCode(); got != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			got,
		)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected recorder status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}
}

func TestResponseWriter_WriteImplicitStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseWriter(rec)

	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != 5 {
		t.Fatalf("expected 5 bytes, got %d", n)
	}

	if w.StatusCode() != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			w.StatusCode(),
		)
	}

	if w.BytesWritten() != 5 {
		t.Fatalf(
			"expected 5 bytes written, got %d",
			w.BytesWritten(),
		)
	}
}

func TestResponseWriter_ReadFrom(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseWriter(rec)

	source := strings.NewReader("hello world")

	n, err := w.ReadFrom(source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != int64(len("hello world")) {
		t.Fatalf(
			"expected %d bytes, got %d",
			len("hello world"),
			n,
		)
	}

	if w.BytesWritten() != len("hello world") {
		t.Fatalf(
			"expected %d bytes written, got %d",
			len("hello world"),
			w.BytesWritten(),
		)
	}

	if rec.Body.String() != "hello world" {
		t.Fatalf(
			"unexpected body: %q",
			rec.Body.String(),
		)
	}
}

func TestResponseWriter_ReadFromUsesUnderlyingReaderFrom(t *testing.T) {
	rec := httptest.NewRecorder()

	underlying := &readerFromWriter{
		ResponseWriter: rec,
	}

	w := NewResponseWriter(underlying)

	_, err := w.ReadFrom(
		strings.NewReader("test"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !underlying.readFromCalled {
		t.Fatal("expected underlying ReaderFrom to be used")
	}

	if rec.Body.String() != "test" {
		t.Fatalf(
			"unexpected body: %q",
			rec.Body.String(),
		)
	}
}

func TestResponseWriter_FlushCapability(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseWriter(rec)

	controller := http.NewResponseController(w)

	if err := controller.Flush(); err != nil {
		t.Fatalf(
			"expected Flush to be supported, got %v",
			err,
		)
	}
}

func TestResponseWriter_Unwrap(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseWriter(rec)

	if got := w.Unwrap(); got != rec {
		t.Fatal("Unwrap returned unexpected ResponseWriter")
	}
}

func TestResponseWriter_HijackCapability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		rw := NewResponseWriter(w)

		// Type assertion выполняется через интерфейс.
		var writer http.ResponseWriter = rw

		hijacker, ok := writer.(http.Hijacker)
		if !ok {
			t.Fatal("ResponseWriter does not preserve http.Hijacker")
		}

		conn, _, err := hijacker.Hijack()
		if err != nil {
			t.Fatalf("Hijack() failed: %v", err)
		}

		defer conn.Close()

		_, err = conn.Write([]byte(
			"HTTP/1.1 200 OK\r\n" +
				"Content-Length: 2\r\n" +
				"Connection: close\r\n" +
				"\r\n" +
				"OK",
		))
		if err != nil {
			t.Fatalf(
				"failed to write hijacked response: %v",
				err,
			)
		}
	}))

	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != "OK" {
		t.Fatalf(
			"unexpected response body: %q",
			string(body),
		)
	}
}

// readerFromWriter используется для проверки,
// что ResponseWriter сохраняет capability io.ReaderFrom.
type readerFromWriter struct {
	http.ResponseWriter

	readFromCalled bool
}

func (w *readerFromWriter) ReadFrom(
	r io.Reader,
) (int64, error) {
	w.readFromCalled = true

	return io.Copy(
		w.ResponseWriter,
		r,
	)
}

// Compile-time capability checks.

var _ http.ResponseWriter = (*ResponseWriter)(nil)
var _ http.Flusher = (*ResponseWriter)(nil)
var _ http.Hijacker = (*ResponseWriter)(nil)
var _ io.ReaderFrom = (*ResponseWriter)(nil)
var _ http.Pusher = (*ResponseWriter)(nil)
