package app

import "github.com/Rowlyge/kuflow/internal/handler"

// Handlers объединяет HTTP-обработчики приложения.
type Handlers struct {
	Health *handler.HealthHandler
}

// NewHandlers создаёт обработчики приложения.
func NewHandlers(
	services *Services,
) (*Handlers, error) {

	return &Handlers{
		Health: handler.NewHealthHandler(
			services.Health,
		),
	}, nil
}
