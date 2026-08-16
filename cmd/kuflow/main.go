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
	"github.com/joho/godotenv"
)

func main() {

	// Загружаем .env, если файл существует.
	// Если его нет — продолжаем работу, используя
	// переменные окружения системы.
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment")
	}

	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL")

	// Запускаем lifecycle background workers.
	//
	// App самостоятельно управляет:
	// - Health Checker
	// - Auth Cache Refresher
	// - Rate Limiter Cleaner
	application.Start()

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

			log.Printf(
				"HTTP server error: %v",
				err,
			)
		}
	}()

	<-ctx.Done()

	log.Println("Shutdown signal received")

	// Сначала останавливаем HTTP server,
	// чтобы прекратить приём новых запросов.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"HTTP server shutdown error: %v",
			err,
		)
	}

	log.Println("HTTP server stopped")

	// Затем App останавливает background workers,
	// ждёт их завершения и закрывает DB pool.
	if err := application.Close(); err != nil {
		log.Fatal(err)
	}

	log.Println("Application stopped")
}
