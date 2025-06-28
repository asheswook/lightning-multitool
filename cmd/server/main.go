package main

import (
	"lmt/internal/server"
	"log/slog"
	"net/http"
)

func bootstrap(errChan chan error) {
	slog.Info("Starting server, listening on :8080")
	errChan <- http.ListenAndServe(":8080", server.Router())
}

func main() {
	errChan := make(chan error)
	go bootstrap(errChan)

	for {
		select {
		case err := <-errChan:
			slog.Error("Server failed to start", "error", err)
			panic(err)
		}
	}
}
