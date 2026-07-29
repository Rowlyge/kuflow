package balancer

import (
	"fmt"
	"sync/atomic"

	"github.com/Rowlyge/kuflow/internal/upstream"
)

// RoundRobin реализует классический алгоритм
// циклического распределения запросов.
type RoundRobin struct {
	manager *upstream.Manager

	counter uint64
}

// NewRoundRobin создаёт балансировщик.
func NewRoundRobin(
	manager *upstream.Manager,
) *RoundRobin {

	return &RoundRobin{
		manager: manager,
	}
}

// Next возвращает следующий доступный upstream.
func (r *RoundRobin) Next() (*upstream.Upstream, error) {

	if r.manager.Count() == 0 {
		return nil, fmt.Errorf("no upstreams available")
	}

	upstreams := r.manager.Upstreams()

	// Пробуем обойти все upstream-серверы.
	for i := 0; i < len(upstreams); i++ {

		index := atomic.AddUint64(
			&r.counter,
			1,
		)

		server := upstreams[int(index-1)%len(upstreams)]

		if server.Alive() {
			return server, nil
		}
	}

	return nil, fmt.Errorf("no healthy upstream available")
}
