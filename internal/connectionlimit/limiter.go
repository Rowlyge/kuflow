package connectionlimit

import "sync"

// Limiter ограничивает количество одновременно активных
// соединений для каждого API key.
//
// Limiter не зависит от HTTP и может использоваться
// любым слоем приложения.
type Limiter struct {
	mu sync.Mutex

	max int

	active map[string]int
}

// New создаёт Connection Limiter.
//
// max определяет максимальное количество одновременно
// активных соединений для одного API key.
func New(max int) *Limiter {
	if max <= 0 {
		max = 1
	}

	return &Limiter{
		max:    max,
		active: make(map[string]int),
	}
}

// Acquire пытается занять один connection slot.
//
// Возвращает true, если connection разрешён.
// Возвращает false, если для API key уже достигнут лимит.
func (l *Limiter) Acquire(apiKey string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.active[apiKey] >= l.max {
		return false
	}

	l.active[apiKey]++

	return true
}

// Release освобождает один connection slot.
//
// Если для API key больше нет активных соединений,
// его запись удаляется из внутреннего хранилища.
func (l *Limiter) Release(apiKey string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	active := l.active[apiKey]

	if active <= 1 {
		delete(l.active, apiKey)
		return
	}

	l.active[apiKey] = active - 1
}

// Active возвращает количество активных соединений
// для указанного API key.
func (l *Limiter) Active(apiKey string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.active[apiKey]
}
