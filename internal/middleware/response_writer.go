package middleware

import "net/http"

// ResponseWriter сохраняет HTTP-статус,
// отправленный клиенту.
type ResponseWriter struct {
	http.ResponseWriter

	statusCode int
}

// WriteHeader перехватывает статус ответа.
func (w *ResponseWriter) WriteHeader(statusCode int) {

	w.statusCode = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}

// StatusCode возвращает HTTP-статус.
func (w *ResponseWriter) StatusCode() int {

	if w.statusCode == 0 {
		return http.StatusOK
	}

	return w.statusCode
}
