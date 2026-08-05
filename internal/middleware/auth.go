package middleware

import (
	"net/http"

	authservice "github.com/Rowlyge/kuflow/internal/service/auth"
)

const HeaderAPIKey = "X-API-Key"

// NewAuth создаёт middleware проверки API Key.
func NewAuth(
	service *authservice.Service,
) Middleware {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			apiKey := r.Header.Get(HeaderAPIKey)

			err := service.Validate(
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

			next.ServeHTTP(
				w,
				r,
			)
		})
	}
}
