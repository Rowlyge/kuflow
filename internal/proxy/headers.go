package proxy

import (
	"net"
	"net/http"
)

// addForwardHeaders добавляет стандартные заголовки Reverse Proxy.
func addForwardHeaders(r *http.Request) {

	// X-Forwarded-Host сообщает целевой системе,
	// к какому хосту первоначально обращался клиент.
	r.Header.Set(
		"X-Forwarded-Host",
		r.Host,
	)

	// X-Forwarded-Proto сообщает протокол.
	if r.TLS == nil {
		r.Header.Set(
			"X-Forwarded-Proto",
			"http",
		)
	} else {
		r.Header.Set(
			"X-Forwarded-Proto",
			"https",
		)
	}

	// Если известен IP клиента — сохраняем его.
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {

		prior := r.Header.Get("X-Forwarded-For")

		if prior == "" {
			r.Header.Set(
				"X-Forwarded-For",
				ip,
			)
		} else {
			r.Header.Set(
				"X-Forwarded-For",
				prior+", "+ip,
			)
		}
	}
}
