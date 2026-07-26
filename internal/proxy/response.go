package proxy

import "net/http"

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

		return nil
	}
}
