package balancer

import "github.com/Rowlyge/kuflow/internal/upstream"

// Balancer определяет алгоритм выбора
// следующего upstream-сервера.
type Balancer interface {

	// Next возвращает upstream,
	// который должен обработать запрос.
	Next() (*upstream.Upstream, error)
}
