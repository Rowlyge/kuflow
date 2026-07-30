package clientip

import (
	"net"
	"net/http"
	"strings"
)

func Get(
	r *http.Request,
) string {

	if forwarded := r.Header.Get(
		"X-Forwarded-For",
	); forwarded != "" {

		return strings.TrimSpace(
			strings.Split(
				forwarded,
				",",
			)[0],
		)
	}

	if realIP := r.Header.Get(
		"X-Real-IP",
	); realIP != "" {

		return realIP
	}

	host, _, err := net.SplitHostPort(
		r.RemoteAddr,
	)

	if err == nil {
		return host
	}

	return r.RemoteAddr
}
