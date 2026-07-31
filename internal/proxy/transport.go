package proxy

import (
	"net"
	"net/http"
	"time"

	"github.com/Rowlyge/kuflow/internal/config"
)

// newTransport создаёт HTTP Transport,
// используемый Reverse Proxy.
func newTransport(
	cfg config.ProxyConfig,
) *http.Transport {

	return &http.Transport{

		DialContext: (&net.Dialer{

			Timeout: cfg.DialTimeout,

			// Обычно KeepAlive редко меняют,
			// поэтому пока оставляем константой.
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// Это уже параметры производительности,
		// а не конфигурации приложения.
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,

		IdleConnTimeout: 90 * time.Second,

		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
	}
}
