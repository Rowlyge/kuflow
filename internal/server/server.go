package server

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/router"
)

// Server управляет HTTP-сервером приложения.
type Server struct {
	httpServer *http.Server

	mu       sync.RWMutex
	listener net.Listener
}

// New создаёт HTTP-сервер приложения.
func New(
	application *app.App,
) *Server {

	mux := router.New(application)

	return newWithHandler(
		mux,
		application.Config.Server,
	)
}

// newWithHandler создаёт HTTP-сервер с переданным handler.
//
// Используется для тестирования HTTP lifecycle без необходимости
// создавать полный App с production-зависимостями.
func newWithHandler(
	handler http.Handler,
	cfg config.ServerConfig,
) *Server {

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: handler,

			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Start запускает HTTP-сервер.
//
// Listener создаётся непосредственно перед запуском Serve,
// поэтому тесты могут использовать динамический порт :0
// и получить фактический адрес после запуска.
func (s *Server) Start() error {

	listener, err := net.Listen(
		"tcp",
		s.httpServer.Addr,
	)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	return s.httpServer.Serve(listener)
}

// Shutdown корректно завершает работу HTTP-сервера.
func (s *Server) Shutdown(
	ctx context.Context,
) error {

	return s.httpServer.Shutdown(ctx)
}

// Addr возвращает фактический адрес HTTP listener.
//
// Используется преимущественно в тестах после успешного запуска.
func (s *Server) Addr() net.Addr {

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return nil
	}

	return s.listener.Addr()
}
