package health

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

// Checker периодически проверяет доступность upstream-серверов.
type Checker struct {
	manager *upstream.Manager

	client *http.Client

	interval time.Duration
	path     string

	ctx    context.Context
	cancel context.CancelFunc
}

// NewChecker создаёт Health Checker.
func NewChecker(
	manager *upstream.Manager,
	cfg config.HealthConfig,
) *Checker {

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	return &Checker{

		manager: manager,

		ctx:    ctx,
		cancel: cancel,

		interval: cfg.Interval,
		path:     cfg.Path,

		client: &http.Client{

			Timeout: cfg.Timeout,

			Transport: &http.Transport{

				DialContext: (&net.Dialer{
					Timeout: cfg.Timeout,
				}).DialContext,
			},
		},
	}
}

// Start запускает периодическую проверку серверов.
func (c *Checker) Start() {

	go func() {

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		// Выполняем первую проверку сразу после запуска.
		c.CheckAll()

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

// Stop завершает работу Health Checker.
func (c *Checker) Stop() {
	c.cancel()
}

// CheckAll проверяет все upstream-серверы.
func (c *Checker) CheckAll() {

	for _, server := range c.manager.Upstreams() {
		_ = c.Check(server)
	}
}

// Check проверяет один upstream.
func (c *Checker) Check(
	u *upstream.Upstream,
) error {

	resp, err := c.client.Get(
		u.URL.String() + c.path,
	)

	alive := err == nil &&
		resp != nil &&
		resp.StatusCode == http.StatusOK

	if resp != nil {
		defer resp.Body.Close()
	}

	// Логируем только реальные изменения состояния.
	if u.UpdateAlive(alive) {

		state := "DOWN"
		if alive {
			state = "UP"
		}

		log.Printf(
			"HealthChecker: %s changed state -> %s",
			u.Name,
			state,
		)
	}

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"unexpected status: %d",
			resp.StatusCode,
		)
	}

	return nil
}
