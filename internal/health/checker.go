package health

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Rowlyge/kuflow/internal/upstream"
)

const checkInterval = 5 * time.Second

// Checker периодически проверяет доступность upstream-серверов.
type Checker struct {
	manager *upstream.Manager

	// HTTP-клиент используется
	// для проверки upstream-серверов.
	client *http.Client

	ctx    context.Context
	cancel context.CancelFunc
}

// NewChecker создаёт Health Checker.
func NewChecker(
	manager *upstream.Manager,
) *Checker {

	ctx, cancel := context.WithCancel(context.Background())

	return &Checker{
		manager: manager,
		ctx:     ctx,
		cancel:  cancel,

		client: &http.Client{
			Timeout: 2 * time.Second,

			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 2 * time.Second,
				}).DialContext,
			},
		},
	}
}

// Start запускает периодическую проверку серверов.
func (c *Checker) Start() {

	go func() {

		ticker := time.NewTicker(checkInterval)
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
		u.URL.String() + "/health",
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
