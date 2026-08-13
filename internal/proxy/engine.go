package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
)

// Engine управляет жизненным циклом Proxy.
type Engine struct {

	// Алгоритм выбора upstream.
	balancer balancer.Balancer

	// Общий HTTP Transport.
	transport http.RoundTripper

	// Настроенный Reverse Proxy.
	proxy *httputil.ReverseProxy
}

// NewEngine создаёт Proxy Engine.
func NewEngine(
	b balancer.Balancer,
	cfg config.ProxyConfig,
) (*Engine, error) {

	transport := newTransport(cfg)

	rp := &httputil.ReverseProxy{
		Rewrite: newDirector(b),

		Transport: transport,

		ModifyResponse: newModifyResponse(),

		ErrorHandler: newErrorHandler(),

		FlushInterval: cfg.FlushInterval,
	}

	return &Engine{

		balancer: b,

		transport: transport,

		proxy: rp,
	}, nil
}

// ServeHTTP реализует http.Handler.
func (e *Engine) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {

	switch DetectMode(r) {

	case ModeReverse:
		e.proxy.ServeHTTP(w, r)

	case ModeHTTPProxy:
		e.serveForward(
			w,
			r,
		)

	case ModeCONNECT:
		e.serveCONNECT(
			w,
			r,
		)

	default:
		http.Error(
			w,
			"Unknown proxy mode",
			http.StatusBadRequest,
		)
	}
}
