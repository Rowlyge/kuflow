package middleware

import "net/http"

// Middleware описывает функцию,
// которая оборачивает HTTP-обработчик.
type Middleware func(http.Handler) http.Handler

// Chain объединяет несколько middleware
// в одну цепочку.
func Chain(
	handler http.Handler,
	middlewares ...Middleware,
) http.Handler {

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

// Default возвращает стандартную
// цепочку middleware KuFlow.
func Default(
	handler http.Handler,
	middlewares ...Middleware,
) http.Handler {

	return Chain(
		handler,
		middlewares...,
	)
}
