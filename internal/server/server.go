package server

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/router"
)

// Start запускает HTTP-сервер приложения.
func Start(application *app.App) error {

	mux := router.New(application)

	return http.ListenAndServe(
		":"+application.Config.Server.Port,
		mux,
	)
}
