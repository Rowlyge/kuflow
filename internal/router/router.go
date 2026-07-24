package router

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/handler"
)

func New(app *app.App) *http.ServeMux {

	mux := http.NewServeMux()

	healthHandler := handler.NewHealthHandler(
		app.Services.Health,
	)

	mux.HandleFunc(
		"/health",
		healthHandler.GetStatus,
	)

	return mux
}
