package server

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

// API provides an HTTP server for administrative tasks, like stopping the application.
type API struct{}

// NewAPI creates a new API server instance.
func NewAPI() *API {
	return &API{}
}

// ListenAndServe starts the API server on the given address.
func (a *API) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/stop", a.stop)
	slog.Info("Starting API server", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

// stop handles the /api/stop request, shutting down the application.
func (a *API) stop(w http.ResponseWriter, req *http.Request) {
	slog.Info("Received stop request from API")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Application is shutting down..."))

	go func() {
		time.Sleep(1 * time.Second) // Give response time to send
		os.Exit(0)
	}()
}
