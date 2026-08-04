package proxy

import (
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const connectIdleTimeout = 5 * time.Minute

// serveCONNECT обрабатывает HTTPS CONNECT.
func (e *Engine) serveCONNECT(
	w http.ResponseWriter,
	r *http.Request,
) {

	targetConn, err := net.Dial(
		"tcp",
		r.Host,
	)
	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusBadGateway,
		)

		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {

		targetConn.Close()

		http.Error(
			w,
			"Hijacking not supported",
			http.StatusInternalServerError,
		)

		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {

		targetConn.Close()

		return
	}

	_, _ = clientConn.Write(
		[]byte("HTTP/1.1 200 Connection Established\r\n\r\n"),
	)

	var uploaded int64
	var downloaded int64

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		transfer(
			targetConn,
			clientConn,
			&uploaded,
		)
	}()

	go func() {
		defer wg.Done()

		transfer(
			clientConn,
			targetConn,
			&downloaded,
		)
	}()

	wg.Wait()

	_ = clientConn.Close()
	_ = targetConn.Close()

	log.Printf(
		"CONNECT closed %s uploaded=%dB downloaded=%dB",
		r.Host,
		uploaded,
		downloaded,
	)
}

// transfer копирует поток данных.
func transfer(
	dst net.Conn,
	src net.Conn,
	counter *int64,
) {

	buffer := make([]byte, 32*1024)

	for {

		_ = src.SetReadDeadline(
			time.Now().Add(connectIdleTimeout),
		)

		n, err := src.Read(buffer)

		if n > 0 {

			_ = dst.SetWriteDeadline(
				time.Now().Add(connectIdleTimeout),
			)

			written, writeErr := dst.Write(
				buffer[:n],
			)

			atomic.AddInt64(
				counter,
				int64(written),
			)

			if writeErr != nil {
				break
			}
		}

		if err != nil {

			if ne, ok := err.(net.Error); ok && ne.Timeout() {

				log.Printf(
					"CONNECT idle timeout",
				)
			}

			break
		}
	}
}
