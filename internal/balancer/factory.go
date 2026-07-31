package balancer

import (
	"fmt"

	"github.com/Rowlyge/kuflow/internal/upstream"
)

// New создаёт балансировщик по имени алгоритма.
func New(
	algorithm string,
	manager *upstream.Manager,
) (Balancer, error) {

	switch algorithm {

	case "round_robin":
		return NewRoundRobin(manager), nil

	default:
		return nil, fmt.Errorf(
			"unknown balancer: %s",
			algorithm,
		)
	}
}
