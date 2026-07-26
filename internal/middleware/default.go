package middleware

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/service"
)

// Default возвращает стандартную цепочку middleware.
func Default(
	handler http.Handler,
	telemetryService *service.TelemetryService,
) http.Handler {

	return Chain(

		handler,

		Recovery,
		RequestID,

		NewLogger(),

		NewTelemetry(
			telemetryService,
		),
	)
}
