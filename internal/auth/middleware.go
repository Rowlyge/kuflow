package auth

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/config"
)

type Middleware struct {
	validator *Validator

	header string
}

func NewMiddleware(
	validator *Validator,
	cfg config.AuthConfig,
) *Middleware {

	return &Middleware{

		validator: validator,

		header: cfg.APIKeyHeader,
	}
}

func (m *Middleware) Wrap(
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		apiKey := r.Header.Get(m.header)

		if err := m.validator.Validate(
			r.Context(),
			apiKey,
		); err != nil {

			http.Error(
				w,
				"Unauthorized",
				http.StatusUnauthorized,
			)

			return
		}

		next.ServeHTTP(
			w,
			r,
		)
	})
}
