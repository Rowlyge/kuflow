package metrics

import (
	"fmt"
	"io"
	"time"
)

// WritePrometheus записывает метрики
// в текстовом формате Prometheus.
func (c *Collector) WritePrometheus(
	w io.Writer,
) {

	s := c.Snapshot()

	// ==========================
	// HTTP Requests
	// ==========================

	fmt.Fprintf(w,
		"# HELP kuflow_requests_total Total HTTP requests.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_requests_total counter\n",
	)
	fmt.Fprintf(w,
		"kuflow_requests_total %d\n\n",
		s.HTTP.Requests,
	)

	fmt.Fprintf(w,
		"# HELP kuflow_requests_success_total Successful HTTP requests.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_requests_success_total counter\n",
	)
	fmt.Fprintf(w,
		"kuflow_requests_success_total %d\n\n",
		s.HTTP.Success,
	)

	fmt.Fprintf(w,
		"# HELP kuflow_requests_failed_total Failed HTTP requests.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_requests_failed_total counter\n",
	)
	fmt.Fprintf(w,
		"kuflow_requests_failed_total %d\n\n",
		s.HTTP.Failed,
	)

	// ==========================
	// Traffic
	// ==========================

	fmt.Fprintf(w,
		"# HELP kuflow_bytes_in_total Total incoming bytes.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_bytes_in_total counter\n",
	)
	fmt.Fprintf(w,
		"kuflow_bytes_in_total %d\n\n",
		s.HTTP.BytesIn,
	)

	fmt.Fprintf(w,
		"# HELP kuflow_bytes_out_total Total outgoing bytes.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_bytes_out_total counter\n",
	)
	fmt.Fprintf(w,
		"kuflow_bytes_out_total %d\n\n",
		s.HTTP.BytesOut,
	)

	// ==========================
	// Latency
	// ==========================

	fmt.Fprintf(w,
		"# HELP kuflow_latency_average_ms Average request latency.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_latency_average_ms gauge\n",
	)
	fmt.Fprintf(w,
		"kuflow_latency_average_ms %.3f\n\n",
		s.HTTP.AverageDurationMs,
	)

	fmt.Fprintf(w,
		"# HELP kuflow_latency_max_ms Maximum request latency.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_latency_max_ms gauge\n",
	)
	fmt.Fprintf(w,
		"kuflow_latency_max_ms %.3f\n\n",
		s.HTTP.MaxDurationMs,
	)

	// ==========================
	// Telemetry
	// ==========================

	fmt.Fprintf(w,
		"# HELP kuflow_telemetry_saved_total Saved telemetry rows.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_telemetry_saved_total counter\n",
	)
	fmt.Fprintf(w,
		"kuflow_telemetry_saved_total %d\n\n",
		s.Telemetry.Saved,
	)

	fmt.Fprintf(w,
		"# HELP kuflow_telemetry_failed_total Failed telemetry writes.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_telemetry_failed_total counter\n",
	)
	fmt.Fprintf(w,
		"kuflow_telemetry_failed_total %d\n\n",
		s.Telemetry.Failed,
	)

	// ==========================
	// Runtime
	// ==========================

	uptime := time.Since(c.StartedAt()).Seconds()

	fmt.Fprintf(w,
		"# HELP kuflow_uptime_seconds Process uptime.\n",
	)
	fmt.Fprintf(w,
		"# TYPE kuflow_uptime_seconds gauge\n",
	)
	fmt.Fprintf(w,
		"kuflow_uptime_seconds %.0f\n",
		uptime,
	)
}
