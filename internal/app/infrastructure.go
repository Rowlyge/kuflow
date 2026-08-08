package app

import (
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/health"
	"github.com/Rowlyge/kuflow/internal/metrics"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

// Infrastructure объединяет инфраструктурные компоненты приложения.
type Infrastructure struct {

	// Менеджер upstream-серверов.
	Upstreams *upstream.Manager

	// Health Checker.
	HealthChecker *health.Checker

	// Collector хранит все runtime-метрики приложения.
	Collector *metrics.Collector
}

// NewInfrastructure создаёт инфраструктуру приложения.
func NewInfrastructure(
	cfg config.ProxyConfig,
	healthCfg config.HealthConfig,
) (*Infrastructure, error) {

	manager, err := upstream.NewManager(
		cfg.Upstreams,
	)
	if err != nil {
		return nil, err
	}

	// Единый Collector для всех инфраструктурных компонентов.
	collector := metrics.NewCollector()

	// Health Checker использует тот же Collector,
	// чтобы результаты active health checks попадали
	// в общую систему runtime-метрик.
	healthChecker := health.NewChecker(
		manager,
		healthCfg,
		collector,
	)

	return &Infrastructure{

		Upstreams: manager,

		HealthChecker: healthChecker,

		Collector: collector,
	}, nil

}
