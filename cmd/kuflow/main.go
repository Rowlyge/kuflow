package main

import (
	"log"

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

	defer application.Close()

	log.Println("Connected to PostgreSQL")

	if err := server.Start(application); err != nil {
		log.Fatal(err)
	}
}
