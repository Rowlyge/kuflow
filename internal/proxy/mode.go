package proxy

import "net/http"

// Mode определяет режим работы прокси.
type Mode int

const (
	ModeReverse Mode = iota
	ModeHTTPProxy
	ModeCONNECT
)

// DetectMode определяет тип входящего запроса.
func DetectMode(r *http.Request) Mode {

	switch {

	case r.Method == http.MethodConnect:
		return ModeCONNECT

	case r.URL.IsAbs():
		return ModeHTTPProxy

	default:
		return ModeReverse
	}
}
