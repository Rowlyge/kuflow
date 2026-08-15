package middleware

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

// ResponseWriter оборачивает стандартный
// http.ResponseWriter и позволяет получить
// информацию об отправленном ответе.
//
// Wrapper сохраняет дополнительные capabilities
// underlying ResponseWriter через http.ResponseController.
type ResponseWriter struct {
	http.ResponseWriter

	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

// NewResponseWriter создаёт обёртку
// над стандартным ResponseWriter.
func NewResponseWriter(
	w http.ResponseWriter,
) *ResponseWriter {

	return &ResponseWriter{
		ResponseWriter: w,

		// Если WriteHeader не вызывался,
		// net/http считает статус ответа равным 200.
		statusCode: http.StatusOK,
	}
}

// Unwrap возвращает underlying ResponseWriter.
//
// Это позволяет http.ResponseController пройти
// через middleware-wrapper и получить доступ
// к дополнительным возможностям исходного writer.
func (w *ResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// WriteHeader сохраняет HTTP-статус
// перед отправкой ответа клиенту.
//
// Как и net/http, учитываем только первый
// вызов WriteHeader.
func (w *ResponseWriter) WriteHeader(
	status int,
) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.statusCode = status

	w.ResponseWriter.WriteHeader(status)
}

// Write записывает тело ответа клиенту
// и считает количество отправленных байт.
func (w *ResponseWriter) Write(
	data []byte,
) (int, error) {

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(data)

	w.bytesWritten += n

	return n, err
}

// Flush сохраняет поддержку http.Flusher.
func (w *ResponseWriter) Flush() {
	_ = http.NewResponseController(
		w.ResponseWriter,
	).Flush()
}

// FlushError сохраняет поддержку FlushError,
// если underlying writer её предоставляет.
func (w *ResponseWriter) FlushError() error {
	return http.NewResponseController(
		w.ResponseWriter,
	).Flush()
}

// Hijack сохраняет поддержку HTTP connection hijacking.
//
// Это критично для CONNECT.
func (w *ResponseWriter) Hijack() (
	net.Conn,
	*bufio.ReadWriter,
	error,
) {
	return http.NewResponseController(
		w.ResponseWriter,
	).Hijack()
}

// SetReadDeadline сохраняет поддержку
// deadline для underlying connection.
func (w *ResponseWriter) SetReadDeadline(
	deadline time.Time,
) error {
	return http.NewResponseController(
		w.ResponseWriter,
	).SetReadDeadline(deadline)
}

// SetWriteDeadline сохраняет поддержку
// deadline для underlying connection.
func (w *ResponseWriter) SetWriteDeadline(
	deadline time.Time,
) error {
	return http.NewResponseController(
		w.ResponseWriter,
	).SetWriteDeadline(deadline)
}

// EnableFullDuplex сохраняет поддержку
// full-duplex HTTP/1.x handling.
func (w *ResponseWriter) EnableFullDuplex() error {
	return http.NewResponseController(
		w.ResponseWriter,
	).EnableFullDuplex()
}

// ReadFrom сохраняет оптимизированный io.Copy path,
// если underlying ResponseWriter реализует io.ReaderFrom.
func (w *ResponseWriter) ReadFrom(
	r io.Reader,
) (int64, error) {

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if readerFrom, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := readerFrom.ReadFrom(r)

		w.bytesWritten += int(n)

		return n, err
	}

	n, err := io.Copy(w.ResponseWriter, r)

	w.bytesWritten += int(n)

	return n, err
}

// Push сохраняет поддержку HTTP/2 server push
// для underlying writer.
//
// HTTP/2 server push deprecated, но capability
// не следует терять на уровне wrapper.
func (w *ResponseWriter) Push(
	target string,
	opts *http.PushOptions,
) error {

	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}

	return errors.New("http push is not supported")
}

// StatusCode возвращает HTTP-статус ответа.
func (w *ResponseWriter) StatusCode() int {
	return w.statusCode
}

// BytesWritten возвращает количество
// отправленных клиенту байт.
func (w *ResponseWriter) BytesWritten() int {
	return w.bytesWritten
}
