package health

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Rowlyge/kuflow/internal/upstream"
)

// Checker отвечает за проверку доступности upstream-серверов.
type Checker struct {
	client *http.Client
}

// NewChecker создаёт Health Checker.
func NewChecker() *Checker {

	return &Checker{
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

// Check проверяет доступность одного upstream.
func (c *Checker) Check(
	u *upstream.Upstream,
) error {

	resp, err := c.client.Get(
		u.URL.String() + "/health",
	)

	if err != nil {

		u.SetAlive(false)

		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		u.SetAlive(false)

		return fmt.Errorf(
			"unexpected status: %d",
			resp.StatusCode,
		)
	}

	u.SetAlive(true)

	return nil
}

// CheckAll проверяет все upstream-серверы.
func (c *Checker) CheckAll(
	manager *upstream.Manager,
) {

	for _, server := range manager.Upstreams() {

		_ = c.Check(server)
	}
}
