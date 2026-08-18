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

const shutdownTimeout = 10 * time.Second

type serverLifecycle interface {
	Shutdown(context.Context) error
}

type appLifecycle interface {
	Close(context.Context) error
}

// shutdown выполняет общий lifecycle shutdown приложения.
//
// HTTP server останавливается первым, чтобы прекратить приём новых
// запросов. После этого завершается App вместе с background workers
// и остальными ресурсами приложения.
//
// Для server и App используются независимые shutdown contexts.
// Каждый этап имеет собственный timeout.
func shutdown(
	httpServer serverLifecycle,
	application appLifecycle,
	serverShutdownTimeout time.Duration,
	appShutdownTimeout time.Duration,
) error {

	httpShutdownCtx, httpShutdownCancel := context.WithTimeout(
		context.Background(),
		serverShutdownTimeout,
	)
	defer httpShutdownCancel()

	var shutdownErr error

	if err := httpServer.Shutdown(httpShutdownCtx); err != nil {
		shutdownErr = err

		log.Printf(
			"HTTP server shutdown error: %v",
			err,
		)
	}

	log.Println("HTTP server stopped")

	appShutdownCtx, appShutdownCancel := context.WithTimeout(
		context.Background(),
		appShutdownTimeout,
	)
	defer appShutdownCancel()

	if err := application.Close(appShutdownCtx); err != nil {
		log.Printf(
			"Application shutdown error: %v",
			err,
		)

		if shutdownErr != nil {
			return errors.Join(shutdownErr, err)
		}

		return err
	}

	log.Println("Application stopped")

	return shutdownErr
}

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

	if err := shutdown(
		httpServer,
		application,
		shutdownTimeout,
		shutdownTimeout,
	); err != nil {

		log.Printf(
			"Shutdown completed with error: %v",
			err,
		)

		return
	}
}
