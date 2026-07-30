package server

import (
	"context"
	"net/http"
	"time"

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

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + application.Config.Server.Port,
			Handler: mux,

			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
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
