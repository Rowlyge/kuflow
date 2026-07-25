package proxy

import "net/http"

// hopHeaders — заголовки, которые нельзя передавать upstream.
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

// removeHopHeaders удаляет hop-by-hop заголовки
// согласно RFC 7230.
func removeHopHeaders(h http.Header) {

	for _, header := range hopHeaders {
		h.Del(header)
	}
}
