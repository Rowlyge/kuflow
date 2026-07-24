package server

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/router"
)

func Start(app *app.App) error {
	mux := router.New(app)

	return http.ListenAndServe(
		":"+app.Config.Server.Port,
		mux,
	)
}
