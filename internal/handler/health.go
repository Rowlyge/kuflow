package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Rowlyge/kuflow/internal/service"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(service *service.HealthService) *HealthHandler {
	return &HealthHandler{
		service: service,
	}
}

func (h *HealthHandler) GetStatus(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	response := h.service.Status()

	json.NewEncoder(w).Encode(response)
}
