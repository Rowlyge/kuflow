package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
)

// Engine управляет жизненным циклом Reverse Proxy.
type Engine struct {

	// Алгоритм выбора upstream.
	balancer balancer.Balancer

	// Настроенный Reverse Proxy.
	proxy *httputil.ReverseProxy
}

// NewEngine создаёт Proxy Engine.
func NewEngine(
	b balancer.Balancer,
	cfg config.ProxyConfig,
) (*Engine, error) {

	// Собираем Reverse Proxy вручную,
	// чтобы полностью контролировать его поведение.
	rp := &httputil.ReverseProxy{

		Director: newDirector(b),

		Transport: newTransport(cfg),

		ModifyResponse: newModifyResponse(),

		ErrorHandler: newErrorHandler(),

		FlushInterval: cfg.FlushInterval,

		// Позже здесь появятся:
		//
		// BufferPool
		// Rewrite
	}

	return &Engine{

		balancer: b,

		proxy: rp,
	}, nil
}

// ServeHTTP реализует http.Handler.
func (e *Engine) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	e.proxy.ServeHTTP(w, r)
}
