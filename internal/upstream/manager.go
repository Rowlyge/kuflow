package upstream

import (
	"fmt"
	"net/url"
)

// Manager хранит список доступных upstream-серверов.
type Manager struct {
	upstreams []*Upstream
}

// NewManager создаёт менеджер upstream-серверов.
func NewManager(
	targets []string,
) (*Manager, error) {

	upstreams := make([]*Upstream, 0, len(targets))

	for i, target := range targets {

		u, err := url.Parse(target)
		if err != nil {
			return nil, err
		}

		up := &Upstream{
			Name: fmt.Sprintf("upstream-%d", i+1),
			URL:  u,
		}

		up.SetAlive(true)

		upstreams = append(upstreams, up)
	}

	return &Manager{
		upstreams: upstreams,
	}, nil
}

// Upstreams возвращает список upstream-серверов.
func (m *Manager) Upstreams() []*Upstream {
	return m.upstreams
}

// Count возвращает количество upstream-серверов.
func (m *Manager) Count() int {
	return len(m.upstreams)
}

// Default возвращает первый upstream.
func (m *Manager) Default() (*Upstream, error) {

	if len(m.upstreams) == 0 {
		return nil, fmt.Errorf("no upstreams configured")
	}

	return m.upstreams[0], nil
}
