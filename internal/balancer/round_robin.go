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

	index := atomic.AddUint64(
		&r.counter,
		1,
	)

	return upstreams[int(index-1)%len(upstreams)], nil
}
