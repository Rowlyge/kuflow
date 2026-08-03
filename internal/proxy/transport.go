package proxy

import (
	"net"
	"net/http"

	"github.com/Rowlyge/kuflow/internal/config"
)

// newTransport создаёт HTTP Transport,
// используемый и Reverse Proxy, и Forward Proxy.
func newTransport(
	cfg config.ProxyConfig,
) http.RoundTripper {

	return &http.Transport{

		Proxy: nil,

		DialContext: (&net.Dialer{
			Timeout: cfg.DialTimeout,
		}).DialContext,

		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,

		ForceAttemptHTTP2: true,
	}
}
