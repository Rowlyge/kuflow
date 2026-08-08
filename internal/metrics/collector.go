package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Collector хранит все внутренние метрики KuFlow.
type Collector struct {
	mu sync.RWMutex

	startedAt time.Time

	requestsTotal uint64

	successTotal uint64
	failedTotal  uint64

	telemetrySaved  uint64
	telemetryFailed uint64

	rateLimited uint64

	bytesInTotal  uint64
	bytesOutTotal uint64

	healthChecksTotal         uint64
	healthChecksSuccess       uint64
	healthChecksTransportFail uint64
	healthChecksHTTPFail      uint64
	healthStateChanges        uint64
	healthUpstreamsUp         uint64
	healthUpstreamsDown       uint64

	durationTotal time.Duration
	durationMax   time.Duration
}

// NewCollector создаёт новый Collector.
func NewCollector() *Collector {

	return &Collector{
		startedAt: time.Now(),
	}
}

func (c *Collector) IncTelemetrySaved() {
	atomic.AddUint64(&c.telemetrySaved, 1)
}

func (c *Collector) IncTelemetryFailed() {
	atomic.AddUint64(&c.telemetryFailed, 1)
}

// IncRequests увеличивает количество запросов.
func (c *Collector) IncRequests() {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestsTotal++
}

// IncSuccess увеличивает количество успешных запросов.
func (c *Collector) IncSuccess() {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.successTotal++
}

// IncFailed увеличивает количество неуспешных запросов.
func (c *Collector) IncFailed() {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.failedTotal++
}

// AddBytesIn увеличивает объём входящего трафика.
func (c *Collector) AddBytesIn(
	n uint64,
) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.bytesInTotal += n
}

// AddBytesOut увеличивает объём исходящего трафика.
func (c *Collector) AddBytesOut(
	n uint64,
) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.bytesOutTotal += n
}

// ObserveDuration сохраняет информацию о времени обработки запроса.
func (c *Collector) ObserveDuration(
	d time.Duration,
) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.durationTotal += d

	if d > c.durationMax {
		c.durationMax = d
	}
}

// StartedAt возвращает время запуска Collector.
func (c *Collector) StartedAt() time.Time {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.startedAt
}

// RequestsTotal возвращает количество запросов.
func (c *Collector) RequestsTotal() uint64 {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.requestsTotal
}

// SuccessTotal возвращает количество успешных запросов.
func (c *Collector) SuccessTotal() uint64 {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.successTotal
}

// FailedTotal возвращает количество неуспешных запросов.
func (c *Collector) FailedTotal() uint64 {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.failedTotal
}

// BytesInTotal возвращает объём входящего трафика.
func (c *Collector) BytesInTotal() uint64 {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.bytesInTotal
}

// BytesOutTotal возвращает объём исходящего трафика.
func (c *Collector) BytesOutTotal() uint64 {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.bytesOutTotal
}

// DurationTotal возвращает суммарное время обработки запросов.
func (c *Collector) DurationTotal() time.Duration {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.durationTotal
}

// DurationMax возвращает максимальное время обработки запроса.
func (c *Collector) DurationMax() time.Duration {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.durationMax
}

// TelemetrySaved возвращает количество успешно сохранённых записей телеметрии.
func (c *Collector) TelemetrySaved() uint64 {
	return atomic.LoadUint64(&c.telemetrySaved)
}

// TelemetryFailed возвращает количество ошибок сохранения телеметрии.
func (c *Collector) TelemetryFailed() uint64 {
	return atomic.LoadUint64(&c.telemetryFailed)
}

// IncRateLimited увеличивает количество запросов,
// отклонённых Rate Limiter.
func (c *Collector) IncRateLimited() {
	atomic.AddUint64(
		&c.rateLimited,
		1,
	)
}

// RateLimited возвращает количество
// отклонённых запросов.
func (c *Collector) RateLimited() uint64 {
	return atomic.LoadUint64(
		&c.rateLimited,
	)
}

func (c *Collector) IncHealthChecks() {
	atomic.AddUint64(&c.healthChecksTotal, 1)
}

func (c *Collector) IncHealthChecksSuccess() {
	atomic.AddUint64(&c.healthChecksSuccess, 1)
}

func (c *Collector) IncHealthChecksTransportFailure() {
	atomic.AddUint64(&c.healthChecksTransportFail, 1)
}

func (c *Collector) IncHealthChecksHTTPFailure() {
	atomic.AddUint64(&c.healthChecksHTTPFail, 1)
}

func (c *Collector) IncHealthStateChanges() {
	atomic.AddUint64(&c.healthStateChanges, 1)
}

func (c *Collector) SetHealthUpstreams(
	up uint64,
	down uint64,
) {
	atomic.StoreUint64(
		&c.healthUpstreamsUp,
		up,
	)

	atomic.StoreUint64(
		&c.healthUpstreamsDown,
		down,
	)
}

func (c *Collector) HealthChecksTotal() uint64 {
	return atomic.LoadUint64(
		&c.healthChecksTotal,
	)
}

func (c *Collector) HealthChecksSuccess() uint64 {
	return atomic.LoadUint64(
		&c.healthChecksSuccess,
	)
}

func (c *Collector) HealthChecksTransportFailure() uint64 {
	return atomic.LoadUint64(
		&c.healthChecksTransportFail,
	)
}

func (c *Collector) HealthChecksHTTPFailure() uint64 {
	return atomic.LoadUint64(
		&c.healthChecksHTTPFail,
	)
}

func (c *Collector) HealthStateChanges() uint64 {
	return atomic.LoadUint64(
		&c.healthStateChanges,
	)
}

func (c *Collector) HealthUpstreamsUp() uint64 {
	return atomic.LoadUint64(
		&c.healthUpstreamsUp,
	)
}

func (c *Collector) HealthUpstreamsDown() uint64 {
	return atomic.LoadUint64(
		&c.healthUpstreamsDown,
	)
}
