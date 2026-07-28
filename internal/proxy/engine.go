package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/Rowlyge/kuflow/internal/balancer"
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
) (*Engine, error) {

	// Собираем Reverse Proxy вручную,
	// чтобы полностью контролировать его поведение.
	rp := &httputil.ReverseProxy{

		Director: newDirector(b),

		Transport: newTransport(),

		ModifyResponse: newModifyResponse(),

		ErrorHandler: newErrorHandler(),

		// Позже здесь появятся:
		//
		// BufferPool
		// FlushInterval
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
