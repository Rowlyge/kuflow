package metrics

import "time"

// ==========================
// HTTP
// ==========================

type HTTPSnapshot struct {
	Requests uint64 `json:"requests"`

	Success uint64 `json:"success"`
	Failed  uint64 `json:"failed"`

	BytesIn  uint64 `json:"bytes_in"`
	BytesOut uint64 `json:"bytes_out"`

	AverageDurationMs float64 `json:"average_duration_ms"`
	MaxDurationMs     float64 `json:"max_duration_ms"`

	RateLimited uint64 `json:"rate_limited"`
}

// ==========================
// Telemetry
// ==========================

type TelemetrySnapshot struct {
	Saved  uint64 `json:"saved"`
	Failed uint64 `json:"failed"`
}

// ==========================
// Health
// ==========================

type HealthSnapshot struct {
	ChecksTotal           uint64 `json:"checks_total"`
	ChecksSuccess         uint64 `json:"checks_success"`
	ChecksTransportFailed uint64 `json:"checks_transport_failed"`
	ChecksHTTPFailed      uint64 `json:"checks_http_failed"`
	StateChanges          uint64 `json:"state_changes"`
	UpstreamsUp           uint64 `json:"upstreams_up"`
	UpstreamsDown         uint64 `json:"upstreams_down"`
}

// ==========================
// Runtime
// ==========================

type RuntimeSnapshot struct {
	Uptime string `json:"uptime"`
}

// ==========================
// Full Snapshot
// ==========================

type Snapshot struct {
	HTTP      HTTPSnapshot      `json:"http"`
	Telemetry TelemetrySnapshot `json:"telemetry"`
	Health    HealthSnapshot    `json:"health"`
	Runtime   RuntimeSnapshot   `json:"runtime"`
}

// Snapshot возвращает снимок текущих метрик.
func (c *Collector) Snapshot() Snapshot {

	requests := c.RequestsTotal()

	var avgMs float64

	if requests > 0 {
		avgMs =
			float64(c.DurationTotal().Milliseconds()) /
				float64(requests)
	}

	return Snapshot{

		HTTP: HTTPSnapshot{
			Requests: c.RequestsTotal(),

			Success: c.SuccessTotal(),
			Failed:  c.FailedTotal(),

			BytesIn:  c.BytesInTotal(),
			BytesOut: c.BytesOutTotal(),

			AverageDurationMs: avgMs,

			MaxDurationMs: float64(
				c.DurationMax().Milliseconds(),
			),

			RateLimited: c.RateLimited(),
		},

		Telemetry: TelemetrySnapshot{
			Saved:  c.TelemetrySaved(),
			Failed: c.TelemetryFailed(),
		},

		Health: HealthSnapshot{
			ChecksTotal: c.HealthChecksTotal(),

			ChecksSuccess: c.HealthChecksSuccess(),

			ChecksTransportFailed: c.HealthChecksTransportFailure(),

			ChecksHTTPFailed: c.HealthChecksHTTPFailure(),

			StateChanges: c.HealthStateChanges(),

			UpstreamsUp: c.HealthUpstreamsUp(),

			UpstreamsDown: c.HealthUpstreamsDown(),
		},

		Runtime: RuntimeSnapshot{
			Uptime: time.Since(
				c.StartedAt(),
			).String(),
		},
	}
}
