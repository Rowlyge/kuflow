package proxy

import (
	"log"
	"net/http"
)

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
		upstream := UpstreamFromContext(r.Context())

		log.Printf(
			"Proxy error: err=%v upstream=%v unavailable=%v url=%v",
			err,
			upstream != nil,
			IsUpstreamUnavailable(r.Context()),
			r.URL,
		)

		if IsUpstreamUnavailable(r.Context()) {
			http.Error(
				w,
				"Service Unavailable",
				http.StatusServiceUnavailable,
			)
			return
		}

		if upstream != nil && upstream.Breaker != nil {
			upstream.Breaker.OnFailure()
		}

		http.Error(
			w,
			"Bad Gateway",
			http.StatusBadGateway,
		)
	}
}
