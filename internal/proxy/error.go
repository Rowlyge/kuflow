package proxy

import (
	"log"
	"net/http"
)

// newErrorHandler вызывается,
// если Reverse Proxy не смог обработать запрос.
func newErrorHandler() func(
	http.ResponseWriter,
	*http.Request,
	error,
) {

	return func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {

		// Балансировщик не смог выбрать
		// ни одного доступного upstream.
		if IsUpstreamUnavailable(
			r.Context(),
		) {

			http.Error(
				w,
				"Service Unavailable",
				http.StatusServiceUnavailable,
			)

			return
		}

		// Upstream был выбран, но произошла
		// ошибка при обращении к нему.
		log.Printf(
			"Proxy error: %v",
			err,
		)

		http.Error(
			w,
			"Bad Gateway",
			http.StatusBadGateway,
		)
	}
}
