package app

import (
	"github.com/Rowlyge/kuflow/internal/health"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

// Health объединяет фоновые проверки.
type Health struct {
	Checker *health.Checker
}

// NewHealth создаёт Health-подсистему.
func NewHealth(
	manager *upstream.Manager,
) *Health {

	return &Health{
		Checker: health.NewChecker(manager),
	}
}
