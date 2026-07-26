package proxy

import (
	"net/http"
	"strings"
)

// rewritePath изменяет путь запроса,
// убирая служебный префикс /proxy.
func rewritePath(r *http.Request) {

	path := strings.TrimPrefix(
		r.URL.Path,
		"/proxy",
	)

	if path == "" {
		path = "/"
	}

	r.URL.Path = path

	// RawPath используется,
	// если URL содержит экранированные символы.
	r.URL.RawPath = path
}
