package auth

import (
	"context"
	"net/http"

	"github.com/Rowlyge/kuflow/internal/config"
)

type contextKey string

const APIKeyContextKey contextKey = "api_key"

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
		header:    cfg.APIKeyHeader,
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

		key, err := m.validator.Validate(
			r.Context(),
			apiKey,
		)

		if err != nil {

			http.Error(
				w,
				"Unauthorized",
				http.StatusUnauthorized,
			)

			return
		}

		ctx := context.WithValue(
			r.Context(),
			APIKeyContextKey,
			key,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}
