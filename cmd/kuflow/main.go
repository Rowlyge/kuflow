package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/server"
)

func main() {

	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL")

	httpServer := server.New(application)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {

		log.Printf(
			"HTTP server started on :%s",
			cfg.Server.Port,
		)

		if err := httpServer.Start(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Println("HTTP server stopped")

	if err := application.Close(); err != nil {
		log.Fatal(err)
	}

	log.Println("Application stopped")

}
