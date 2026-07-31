package server

import (
	"context"
	"net/http"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/router"
)

// Server управляет HTTP-сервером приложения.
type Server struct {
	httpServer *http.Server
}

// New создаёт HTTP-сервер.
func New(
	application *app.App,
) *Server {

	mux := router.New(application)

	cfg := application.Config.Server

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: mux,

			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Start запускает HTTP-сервер.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown корректно завершает работу HTTP-сервера.
func (s *Server) Shutdown(
	ctx context.Context,
) error {
	return s.httpServer.Shutdown(ctx)
}
