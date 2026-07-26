package proxy

import (
	"net/http/httputil"
	"net/url"
)

// Engine управляет Reverse Proxy и его настройками.
type Engine struct {
	target *url.URL

	proxy *httputil.ReverseProxy
}

// NewEngine создаёт и настраивает Proxy Engine.
func NewEngine(
	target string,
) (*Engine, error) {

	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(u)

	reverseProxy.Director = newDirector(u)

	reverseProxy.Transport = newTransport()

	reverseProxy.ModifyResponse = newModifyResponse()

	reverseProxy.ErrorHandler = newErrorHandler()

	return &Engine{

		target: u,

		proxy: reverseProxy,
	}, nil
}

// ReverseProxy возвращает настроенный Reverse Proxy.
func (e *Engine) ReverseProxy() *httputil.ReverseProxy {
	return e.proxy
}
