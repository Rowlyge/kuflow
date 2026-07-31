package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Rowlyge/kuflow/internal/metrics"
)

type MetricsHandler struct {
	collector *metrics.Collector
}

func NewMetricsHandler(
	collector *metrics.Collector,
) *MetricsHandler {

	return &MetricsHandler{
		collector: collector,
	}
}

func (h *MetricsHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(
		h.collector.Snapshot(),
	)
}
