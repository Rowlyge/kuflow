package service

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/proxy"
)

// ProxyService отвечает за обработку проксируемых запросов.
type ProxyService struct {
	engine *proxy.Engine
}

// NewProxyService создаёт сервис прокси.
func NewProxyService(
	b balancer.Balancer,
	cfg config.ProxyConfig,
) (*ProxyService, error) {

	engine, err := proxy.NewEngine(
		b,
		cfg,
	)
	if err != nil {
		return nil, err
	}

	return &ProxyService{
		engine: engine,
	}, nil
}

// ServeHTTP делает ProxyService HTTP-обработчиком.
func (s *ProxyService) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.engine.ServeHTTP(w, r)
}
