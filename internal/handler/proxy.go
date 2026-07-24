package handler

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/service"
)

type ProxyHandler struct {
	service *service.ProxyService
}

func NewProxyHandler(service *service.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		service: service,
	}
}

func (h *ProxyHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.service.Proxy().ServeHTTP(w, r)
}
