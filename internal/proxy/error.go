package proxy

import (
	"log"
	"net/http"
)

// newErrorHandler вызывается,
// если запрос к upstream завершился ошибкой.
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
