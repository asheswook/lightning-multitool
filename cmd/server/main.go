package main

import (
	"go.uber.org/dig"
	"lmt/internal/config"
	"lmt/internal/server"
	"lmt/pkg/lndrest"
	"log/slog"
	"net/http"
)

func main() {
	container := dig.New()

	container.Provide(config.NewConfig)
	container.Provide(lndrest.NewClient)

	err := container.Invoke(func(config *config.Config, lndClient *lndrest.Client) {
		slog.Info("Starting server", "addr", config.Server.Host+":"+config.Server.Port)
		if err := http.ListenAndServe(config.Server.Host+":"+config.Server.Port, server.Router(lndClient)); err != nil {
			slog.Error("Server failed to start", "error", err)
			panic(err)
		}
	})

	panic(err)
}
