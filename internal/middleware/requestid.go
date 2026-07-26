package middleware

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/requestid"
)

// NewRequestID добавляет Request ID
// в каждый входящий запрос.
func NewRequestID() Middleware {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			id := requestid.FromHeader(
				r.Header.Get(requestid.HeaderName()),
			)

			r = r.WithContext(
				requestid.IntoContext(
					r.Context(),
					id,
				),
			)

			w.Header().Set(
				requestid.HeaderName(),
				id,
			)

			next.ServeHTTP(
				w,
				r,
			)
		})
	}
}
