package health

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/metrics"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

var (
	// ErrTransportFailure означает,
	// что запрос не смог получить HTTP-ответ.
	ErrTransportFailure = errors.New(
		"health check transport failure",
	)

	// ErrHTTPFailure означает,
	// что upstream ответил HTTP-статусом,
	// который считается неуспешным.
	ErrHTTPFailure = errors.New(
		"health check http failure",
	)
)

// Checker периодически проверяет доступность upstream-серверов.
type Checker struct {
	manager *upstream.Manager

	client *http.Client

	collector *metrics.Collector

	interval time.Duration
	path     string

	ctx    context.Context
	cancel context.CancelFunc
}

// NewChecker создаёт Health Checker.
func NewChecker(
	manager *upstream.Manager,
	cfg config.HealthConfig,
	collector *metrics.Collector,
) *Checker {

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	return &Checker{
		manager: manager,

		client: &http.Client{
			Timeout: cfg.Timeout,

			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: cfg.Timeout,
				}).DialContext,
			},
		},

		collector: collector,

		interval: cfg.Interval,
		path:     cfg.Path,

		ctx:    ctx,
		cancel: cancel,
	}

}

// Start запускает периодическую проверку
// upstream-серверов.
func (c *Checker) Start() {

	go func() {

		// Первая проверка выполняется сразу.
		c.CheckAll()

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {

			case <-ticker.C:
				c.CheckAll()

			case <-c.ctx.Done():
				return
			}
		}
	}()

}

// Stop останавливает Health Checker.
func (c *Checker) Stop() {
	c.cancel()
}

// CheckAll проверяет все upstream-серверы.
func (c *Checker) CheckAll() {

	var up uint64
	var down uint64

	for _, server := range c.manager.Upstreams() {

		if err := c.Check(server); err != nil {

			switch {

			case errors.Is(err, ErrTransportFailure):
				log.Printf(
					"HealthChecker: %s transport failure: %v",
					server.Name,
					err,
				)

			case errors.Is(err, ErrHTTPFailure):
				log.Printf(
					"HealthChecker: %s http failure: %v",
					server.Name,
					err,
				)

			default:
				log.Printf(
					"HealthChecker: %s check failed: %v",
					server.Name,
					err,
				)
			}
		}

		if server.Alive() {
			up++
		} else {
			down++
		}
	}

	if c.collector != nil {
		c.collector.SetHealthUpstreams(
			up,
			down,
		)
	}

}

// Check проверяет доступность одного
// upstream-сервера.
func (c *Checker) Check(
	u *upstream.Upstream,
) error {

	if c.collector != nil {
		c.collector.IncHealthChecks()
	}

	if u == nil {
		return fmt.Errorf(
			"upstream is nil",
		)
	}

	if u.URL == nil {
		return fmt.Errorf(
			"upstream %q has nil URL",
			u.Name,
		)
	}

	target := u.URL.ResolveReference(
		&url.URL{
			Path: c.path,
		},
	)

	req, err := http.NewRequestWithContext(
		c.ctx,
		http.MethodGet,
		target.String(),
		nil,
	)
	if err != nil {

		c.handleFailure(u)

		if c.collector != nil {
			c.collector.IncHealthChecksTransportFailure()
		}

		return fmt.Errorf(
			"%w: %v",
			ErrTransportFailure,
			err,
		)
	}

	resp, err := c.client.Do(req)
	if err != nil {

		c.handleFailure(u)

		if c.collector != nil {
			c.collector.IncHealthChecksTransportFailure()
		}

		return fmt.Errorf(
			"%w: %v",
			ErrTransportFailure,
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {

		c.handleFailure(u)

		if c.collector != nil {
			c.collector.IncHealthChecksHTTPFailure()
		}

		return fmt.Errorf(
			"%w: status %d",
			ErrHTTPFailure,
			resp.StatusCode,
		)
	}

	c.handleSuccess(u)

	if c.collector != nil {
		c.collector.IncHealthChecksSuccess()
	}

	return nil

}

// handleSuccess обрабатывает успешную
// health check-проверку.
func (c *Checker) handleSuccess(
	u *upstream.Upstream,
) {

	if u.UpdateAlive(true) {

		if c.collector != nil {
			c.collector.IncHealthStateChanges()
		}

		log.Printf(
			"HealthChecker: %s changed state -> UP",
			u.Name,
		)
	}

	if u.Breaker != nil {
		u.Breaker.OnSuccess()
	}

}

// handleFailure обрабатывает неуспешную
// health check-проверку.
func (c *Checker) handleFailure(
	u *upstream.Upstream,
) {

	if u.UpdateAlive(false) {

		if c.collector != nil {
			c.collector.IncHealthStateChanges()
		}

		log.Printf(
			"HealthChecker: %s changed state -> DOWN",
			u.Name,
		)
	}

	if u.Breaker != nil {
		u.Breaker.OnFailure()
	}

}
