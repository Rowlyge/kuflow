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
}

// ==========================
// Telemetry
// ==========================

type TelemetrySnapshot struct {
	Saved  uint64 `json:"saved"`
	Failed uint64 `json:"failed"`
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
		},

		Telemetry: TelemetrySnapshot{
			Saved:  c.TelemetrySaved(),
			Failed: c.TelemetryFailed(),
		},

		Runtime: RuntimeSnapshot{
			Uptime: time.Since(
				c.StartedAt(),
			).String(),
		},
	}
}
