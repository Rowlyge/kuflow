package middleware

import "net/http"

// ResponseWriter оборачивает стандартный
// http.ResponseWriter и позволяет получить
// информацию об отправленном ответе.
type ResponseWriter struct {
	http.ResponseWriter

	statusCode   int
	bytesWritten int
}

// NewResponseWriter создаёт обёртку
// над стандартным ResponseWriter.
func NewResponseWriter(
	w http.ResponseWriter,
) *ResponseWriter {

	return &ResponseWriter{
		ResponseWriter: w,

		// По стандарту net/http,
		// если WriteHeader не вызывался,
		// считается, что ответ имеет статус 200.
		statusCode: http.StatusOK,
	}
}

// WriteHeader сохраняет HTTP-статус
// перед отправкой ответа клиенту.
func (w *ResponseWriter) WriteHeader(
	status int,
) {
	w.statusCode = status

	w.ResponseWriter.WriteHeader(status)
}

// Write записывает тело ответа клиенту
// и считает количество отправленных байт.
func (w *ResponseWriter) Write(
	data []byte,
) (int, error) {

	n, err := w.ResponseWriter.Write(data)

	w.bytesWritten += n

	return n, err
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
