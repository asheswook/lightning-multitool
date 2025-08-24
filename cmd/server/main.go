package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/nbd-wtf/go-nostr/nip19"
	"go.uber.org/dig"
	"lmt/internal/app"
	"lmt/internal/config"
	"lmt/internal/server"
	"lmt/pkg/lndrest"
	"lmt/pkg/oksusu"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func ProvideLNDClient(cfg *config.Config) (*lndrest.Client, error) {
	macaroonPath := cfg.LND.MacaroonPath
	if strings.HasPrefix(macaroonPath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
		macaroonPath = filepath.Join(usr.HomeDir, macaroonPath[2:])
	}

	macaroonBytes, err := os.ReadFile(macaroonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read macaroon file: %w", err)
	}
	macaroon := hex.EncodeToString(macaroonBytes)

	return lndrest.NewClient(cfg.LND.Host, macaroon, "")
}

func ProvideZapMonitor(cfg *config.Config, lndClient *lndrest.Client) app.ZapMonitor {
	_, vpub, err := nip19.Decode(cfg.Nostr.PublicKey)
	if err != nil {
		panic(err)
	}

	_, vpriv, err := nip19.Decode(cfg.Nostr.PrivateKey)
	if err != nil {
		panic(err)
	}

	return app.NewZapMonitor(
		lndClient,
		vpub.(string),
		vpriv.(string),
		cfg.Nostr.Relays,
	)
}

func ProvideLNURLHandler(cfg *config.Config) app.LNURLHandler {
	_, vpub, err := nip19.Decode(cfg.Nostr.PublicKey)
	if err != nil {
		panic(err)
	}

	return app.NewLNURLHandler(
		cfg.General.Username,
		cfg.LNURL.Domain,
		vpub.(string),
		cfg.LNURL.MaxSendableMsat,
		cfg.LNURL.MinSendableMsat,
		cfg.LNURL.CommentAllowed,
	)
}

func ProvideLNURLInvoiceHandler(cfg *config.Config, lndClient *lndrest.Client, zapMonitor app.ZapMonitor) app.LNURLInvoiceHandler {
	_, vpub, err := nip19.Decode(cfg.Nostr.PublicKey)
	if err != nil {
		panic(err)
	}

	return app.NewLNURLInvoiceHandler(
		lndClient,
		zapMonitor,
		cfg.General.Username,
		vpub.(string),
	)
}

func ProvideNostrHandler(cfg *config.Config) app.NostrHandler {
	_, vpub, err := nip19.Decode(cfg.Nostr.PublicKey)
	if err != nil {
		panic(err)
	}

	return app.NewNostrHandler(cfg.General.Username, vpub.(string))
}

func ProvideOksusuHandler(cfg *config.Config, lndClient *lndrest.Client, zapMonitor app.ZapMonitor) app.OksusuHandler {
	_, vpub, err := nip19.Decode(cfg.Nostr.PublicKey)
	if err != nil {
		panic(err)
	}

	return app.NewOksusuHandler(
		cfg.General.Username,
		cfg.Oksusu.Server,
		vpub.(string),
		cfg.LNURL.MaxSendableMsat,
		cfg.LNURL.MinSendableMsat,
		cfg.LNURL.CommentAllowed,
		lndClient,
		zapMonitor,
	)
}

func main() {
	container := dig.New()

	if err := container.Provide(config.NewConfig); err != nil {
		panic(err)
	}

	if err := container.Provide(ProvideLNDClient); err != nil {
		panic(err)
	}

	if err := container.Provide(ProvideZapMonitor); err != nil {
		panic(err)
	}

	if err := container.Provide(ProvideLNURLHandler); err != nil {
		panic(err)
	}

	if err := container.Provide(ProvideLNURLInvoiceHandler); err != nil {
		panic(err)
	}

	if err := container.Provide(ProvideNostrHandler); err != nil {
		panic(err)
	}

	if err := container.Provide(server.NewRouter); err != nil {
		panic(err)
	}

	if err := container.Provide(server.NewAPI); err != nil {
		panic(err)
	}

	if err := container.Provide(ProvideOksusuHandler); err != nil {
		panic(err)
	}

	slog.Info("Test1")

	if err := container.Invoke(func(cfg *config.Config, router server.Router, handler app.OksusuHandler, api *server.API) error {
		slog.Info("Invoke")
		go func() {
			slog.Info("Test2")
			if err := api.ListenAndServe("0.0.0.0" + ":" + cfg.API.Port); err != nil {
				panic(err)
			}

			slog.Info("Test3")
		}()

		if cfg.Oksusu.Enabled {
			// Oksu Connect Mode
			if cfg.Oksusu.Token == "" {
				return fmt.Errorf("oksu.token must be set when oksu.enabled is true")
			}
			slog.Info("Starting in Oksusu Connect mode")
			client := oksusu.NewClient(cfg.Oksusu.Server, cfg.Oksusu.Token, handler)

			// Run in a loop to handle automatic reconnections
			for {
				err := client.ConnectAndServe(context.Background())
				if err != nil {
					slog.Error("Oksu client disconnected with error", "error", err)
				}
				slog.Info("Attempting to reconnect in 10 seconds...")
				time.Sleep(10 * time.Second)
			}
		} else {
			// Standalone Web Server Mode
			slog.Info("Starting in standard web server mode")
			return router.ListenAndServe(cfg.Server.Host + ":" + cfg.Server.Port)
		}
	}); err != nil {
		panic(err)
	}
}
