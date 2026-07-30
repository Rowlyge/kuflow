package app

import (
	"github.com/Rowlyge/kuflow/internal/health"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

// Infrastructure объединяет инфраструктурные компоненты приложения.
type Infrastructure struct {

	// Менеджер upstream-серверов.
	Upstreams *upstream.Manager

	// Health Checker.
	HealthChecker *health.Checker
}

// NewInfrastructure создаёт инфраструктуру приложения.
func NewInfrastructure(
	targets []string,
) (*Infrastructure, error) {

	manager, err := upstream.NewManager(targets)
	if err != nil {
		return nil, err
	}

	return &Infrastructure{
		Upstreams: manager,

		HealthChecker: health.NewChecker(
			manager,
		),
	}, nil
}
