package proxy

import (
	"net/http"
)

// serveForward обрабатывает запросы
// в режиме Forward Proxy.
func (e *Engine) serveForward(
	w http.ResponseWriter,
	r *http.Request,
) {

	// Forward Proxy принимает только абсолютные URL.
	if !r.URL.IsAbs() {

		http.Error(
			w,
			"absolute URL required",
			http.StatusBadRequest,
		)

		return
	}

	// Создаём исходящий запрос.
	outReq, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		r.URL.String(),
		r.Body,
	)
	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusBadGateway,
		)

		return
	}

	// Копируем заголовки клиента.
	outReq.Header = r.Header.Clone()

	// RFC 7230.
	removeHopHeaders(outReq.Header)

	// Выполняем запрос.
	resp, err := e.transport.RoundTrip(outReq)
	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusBadGateway,
		)

		return
	}
	defer resp.Body.Close()

	if err := writeResponse(
		w,
		resp,
	); err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusBadGateway,
		)

		return
	}
}
