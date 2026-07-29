package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {

	name := os.Getenv("SERVER_NAME")
	if name == "" {
		name = "Unknown"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()

	// Главная тестовая ручка.
	mux.HandleFunc("/", func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		fmt.Fprintf(
			w,
			"Hello from %s\n",
			name,
		)
	})

	// Health Check.
	mux.HandleFunc("/health", func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "OK")
	})

	log.Printf(
		"%s started on :%s",
		name,
		port,
	)

	log.Fatal(http.ListenAndServe(
		":"+port,
		mux,
	))
}
