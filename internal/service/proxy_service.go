package service

import (
	"github.com/Rowlyge/kuflow/internal/proxy"
)

type ProxyService struct {
	proxy *proxy.Proxy
}

func NewProxyService(target string) (*ProxyService, error) {

	p, err := proxy.New(target)
	if err != nil {
		return nil, err
	}

	return &ProxyService{
		proxy: p,
	}, nil
}

func (s *ProxyService) Proxy() *proxy.Proxy {
	return s.proxy
}
