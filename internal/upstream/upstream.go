package upstream

import "net/url"

// Upstream описывает целевой сервер,
// на который могут быть направлены запросы.
type Upstream struct {

	// Логическое имя сервера.
	Name string

	// Адрес upstream-сервера.
	URL *url.URL

	// Доступен ли сервер.
	Alive bool
}
