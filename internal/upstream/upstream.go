package upstream

import (
	"net/url"
	"sync/atomic"

	"github.com/Rowlyge/kuflow/internal/breaker"
)

// Upstream описывает один backend-сервер.
type Upstream struct {
	// Логическое имя сервера.
	Name string

	// Адрес сервера.
	URL *url.URL

	// Circuit Breaker сервера.
	Breaker *breaker.Breaker

	// Признак доступности.
	alive atomic.Bool

	// Количество последовательных неуспешных
	// health-check проверок.
	healthFailures atomic.Int64

	// Количество последовательных успешных
	// health-check проверок.
	healthSuccesses atomic.Int64
}

// Alive возвращает текущее состояние сервера.
func (u *Upstream) Alive() bool {
	return u.alive.Load()
}

// SetAlive обновляет состояние сервера.
func (u *Upstream) SetAlive(value bool) {
	u.alive.Store(value)
}

// UpdateAlive обновляет состояние сервера.
//
// Возвращает true, если состояние действительно изменилось.
func (u *Upstream) UpdateAlive(value bool) bool {
	old := u.alive.Load()

	if old == value {
		return false
	}

	u.alive.Store(value)

	return true

}

// IncHealthFailures увеличивает количество
// последовательных неуспешных health-check проверок.
//
// Возвращает новое значение счётчика.
func (u *Upstream) IncHealthFailures() int {
	return int(u.healthFailures.Add(1))
}

// ResetHealthFailures сбрасывает счётчик
// последовательных неуспешных health-check проверок.
func (u *Upstream) ResetHealthFailures() {
	u.healthFailures.Store(0)
}

// HealthFailures возвращает количество
// последовательных неуспешных health-check проверок.
func (u *Upstream) HealthFailures() int {
	return int(u.healthFailures.Load())
}

// IncHealthSuccesses увеличивает количество
// последовательных успешных health-check проверок.
//
// Возвращает новое значение счётчика.
func (u *Upstream) IncHealthSuccesses() int {
	return int(u.healthSuccesses.Add(1))
}

// ResetHealthSuccesses сбрасывает счётчик
// последовательных успешных health-check проверок.
func (u *Upstream) ResetHealthSuccesses() {
	u.healthSuccesses.Store(0)
}

// HealthSuccesses возвращает количество
// последовательных успешных health-check проверок.
func (u *Upstream) HealthSuccesses() int {
	return int(u.healthSuccesses.Load())
}

// Available возвращает,
// можно ли использовать upstream.
func (u *Upstream) Available() bool {
	if !u.Alive() {
		return false
	}

	if u.Breaker != nil && !u.Breaker.Allow() {
		return false
	}

	return true

}
