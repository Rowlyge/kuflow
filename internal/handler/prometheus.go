package handler

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/metrics"
)

// PrometheusHandler экспортирует метрики
// в формате Prometheus.
type PrometheusHandler struct {
	collector *metrics.Collector
}

// NewPrometheusHandler создаёт обработчик
// экспорта метрик Prometheus.
func NewPrometheusHandler(
	collector *metrics.Collector,
) *PrometheusHandler {

	return &PrometheusHandler{
		collector: collector,
	}
}

// ServeHTTP реализует http.Handler.
func (h *PrometheusHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {

	w.Header().Set(
		"Content-Type",
		"text/plain; version=0.0.4; charset=utf-8",
	)

	h.collector.WritePrometheus(w)
}
