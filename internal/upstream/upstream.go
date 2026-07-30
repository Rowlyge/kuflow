package upstream

import (
	"net/url"
	"sync/atomic"
)

// Upstream описывает один backend-сервер.
type Upstream struct {

	// Логическое имя сервера.
	Name string

	// Адрес сервера.
	URL *url.URL

	// Признак доступности.
	alive atomic.Bool
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
// Возвращает true, если состояние изменилось.
func (u *Upstream) UpdateAlive(value bool) bool {

	old := u.alive.Load()

	if old == value {
		return false
	}

	u.alive.Store(value)

	return true
}
