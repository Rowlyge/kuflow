package middleware

import "net/http"

// Middleware описывает функцию, которая получает следующий обработчик
// и возвращает новый обработчик с дополнительной логикой.
type Middleware func(http.Handler) http.Handler

// Chain объединяет несколько middleware в одну цепочку.
//
// Middleware оборачиваются в обратном порядке,
// чтобы первый переданный middleware оказался внешним.
func Chain(
	handler http.Handler,
	middlewares ...Middleware,
) http.Handler {

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
