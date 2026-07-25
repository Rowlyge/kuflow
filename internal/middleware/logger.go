package middleware

import (
	"log"
	"net"
	"net/http"
	"time"
)

// LoggerMiddleware отвечает только за логирование запросов.
type LoggerMiddleware struct{}

// NewLogger создаёт middleware логирования.
func NewLogger() Middleware {

	logger := &LoggerMiddleware{}

	return logger.Handler
}

// Handler выполняется для каждого HTTP-запроса.
func (l *LoggerMiddleware) Handler(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rw := NewResponseWriter(w)

		next.ServeHTTP(rw, r)

		requestID, _ := r.Context().
			Value(RequestIDKey).(string)

		log.Printf(
			"[%s] %s %s -> %d (%v) IP=%s",
			requestID,
			r.Method,
			r.URL.Path,
			rw.StatusCode,
			time.Since(start),
			clientIP(r),
		)
	})
}

// clientIP пытается определить реальный IP клиента.
func clientIP(r *http.Request) string {

	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}

	return host
}
