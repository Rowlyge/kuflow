package service

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/proxy"
)

// ProxyService отвечает за обработку всех проксируемых запросов.
type ProxyService struct {
	engine *proxy.Engine
}

// NewProxyService создаёт сервис прокси.
func NewProxyService(
	target string,
) (*ProxyService, error) {

	engine, err := proxy.NewEngine(target)
	if err != nil {
		return nil, err
	}

	return &ProxyService{
		engine: engine,
	}, nil
}

// ServeHTTP делает ProxyService обычным HTTP-обработчиком.
func (s *ProxyService) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.engine.ReverseProxy().ServeHTTP(w, r)
}
