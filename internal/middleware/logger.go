package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/Rowlyge/kuflow/internal/requestid"
)

// NewLogger создаёт middleware логирования.
func NewLogger() Middleware {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			start := time.Now()

			writer := &ResponseWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(
				writer,
				r,
			)

			duration := time.Since(start)

			log.Printf(
				"[%s] %s %s | status=%d | duration=%s | ip=%s",
				requestid.FromContext(r.Context()),
				r.Method,
				r.URL.Path,
				writer.StatusCode(),
				duration.Round(time.Millisecond),
				r.RemoteAddr,
			)
		})
	}
}
