package proxy

import (
	"net"
	"net/http"
	"time"
)

// newTransport создаёт HTTP Transport,
// который используется Reverse Proxy.
func newTransport() *http.Transport {

	return &http.Transport{

		DialContext: (&net.Dialer{

			Timeout: 5 * time.Second,

			KeepAlive: 30 * time.Second,
		}).DialContext,

		MaxIdleConns: 100,

		MaxIdleConnsPerHost: 20,

		IdleConnTimeout: 90 * time.Second,

		ResponseHeaderTimeout: 10 * time.Second,
	}
}
