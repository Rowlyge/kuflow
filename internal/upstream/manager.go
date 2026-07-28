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
//
// Пока поддерживается один сервер.
// Позже список будет загружаться из конфигурации.
func NewManager(
	target string,
) (*Manager, error) {

	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return &Manager{
		upstreams: []*Upstream{
			{
				Name:  "default",
				URL:   u,
				Alive: true,
			},
		},
	}, nil
}

// Upstreams возвращает список upstream-серверов.
func (m *Manager) Upstreams() []*Upstream {
	return m.upstreams
}

// Count возвращает количество upstream.
func (m *Manager) Count() int {
	return len(m.upstreams)
}

// Default возвращает первый upstream.
//
// Пока используется только в тестах.
func (m *Manager) Default() (*Upstream, error) {

	if len(m.upstreams) == 0 {
		return nil, fmt.Errorf("no upstreams configured")
	}

	return m.upstreams[0], nil
}
