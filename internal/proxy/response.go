package proxy

import (
	"io"
	"net/http"
)

// newModifyResponse вызывается
// после получения ответа от upstream.
func newModifyResponse() func(
	*http.Response,
) error {

	return func(
		resp *http.Response,
	) error {

		resp.Header.Set(
			"X-KuFlow",
			"true",
		)

		if upstream := UpstreamFromContext(resp.Request.Context()); upstream != nil {
			upstream.Breaker.OnSuccess()
		}

		return nil
	}
}

// writeResponse копирует ответ upstream клиенту.
func writeResponse(
	w http.ResponseWriter,
	resp *http.Response,
) error {

	removeHopHeaders(resp.Header)

	for key, values := range resp.Header {

		for _, value := range values {

			w.Header().Add(
				key,
				value,
			)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err := io.Copy(
		w,
		resp.Body,
	)

	return err
}
