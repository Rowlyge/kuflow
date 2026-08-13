package proxy

import (
	"errors"
	"net"
	"net/http"

	"github.com/Rowlyge/kuflow/internal/config"
)

var errUpstreamUnavailable = errors.New(
	"no available upstream",
)

// newTransport создаёт HTTP Transport,
// используемый и Reverse Proxy, и Forward Proxy.
func newTransport(
	cfg config.ProxyConfig,
) http.RoundTripper {

	base := &http.Transport{
		Proxy: nil,

		DialContext: (&net.Dialer{
			Timeout: cfg.DialTimeout,
		}).DialContext,

		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,

		ForceAttemptHTTP2: true,
	}

	return &proxyTransport{
		base: base,
	}
}

type proxyTransport struct {
	base http.RoundTripper
}

func (t *proxyTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {

	if IsUpstreamUnavailable(req.Context()) {
		return nil, errUpstreamUnavailable
	}

	return t.base.RoundTrip(req)
}
