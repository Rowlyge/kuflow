package breaker

import (
	"sync"
	"time"
)

// Breaker реализует алгоритм Circuit Breaker.
type Breaker struct {
	mu sync.Mutex

	state State

	config Config

	failures int

	successes int

	openedAt time.Time
}

// New создаёт новый Circuit Breaker.
func New(
	cfg Config,
) *Breaker {

	return &Breaker{

		state: Closed,

		config: cfg,
	}
}

// Allow определяет,
// можно ли отправлять запрос.
func (b *Breaker) Allow() bool {

	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {

	case Closed:

		return true

	case Open:

		if time.Since(
			b.openedAt,
		) >= b.config.OpenTimeout {

			b.state = HalfOpen

			b.successes = 0

			return true
		}

		return false

	case HalfOpen:

		// Пока разрешаем тестовый запрос.
		return true

	default:

		return false
	}
}

// OnSuccess вызывается
// после успешного ответа upstream.
func (b *Breaker) OnSuccess() {

	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {

	case Closed:

		b.failures = 0

	case HalfOpen:

		b.successes++

		if b.successes >= b.config.SuccessThreshold {

			b.state = Closed

			b.failures = 0

			b.successes = 0
		}
	}
}

// OnFailure вызывается
// после ошибки upstream.
func (b *Breaker) OnFailure() {

	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {

	case Closed:

		b.failures++

		if b.failures >= b.config.FailureThreshold {

			b.state = Open

			b.openedAt = time.Now()

			b.failures = 0

			b.successes = 0
		}

	case HalfOpen:

		b.state = Open

		b.openedAt = time.Now()

		b.successes = 0
	}
}

// State возвращает
// текущее состояние Breaker.
func (b *Breaker) State() State {

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.state
}
