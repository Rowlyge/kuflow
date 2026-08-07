package breaker

// State описывает текущее состояние Circuit Breaker.
type State int

const (

	// Closed — нормальная работа.
	Closed State = iota

	// Open — запросы запрещены.
	Open

	// HalfOpen — разрешён тестовый запрос.
	HalfOpen
)
