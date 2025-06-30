package main

import (
	"encoding/hex"
	"fmt"
	"go.uber.org/dig"
	"lmt/internal/config"
	"lmt/internal/server"
	"lmt/pkg/lndrest"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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

func ProvideLNURLHandler(cfg *config.Config) server.LNURLHandler {
	return server.NewLNURLHandler(
		cfg.Username,
		cfg.LNURL.Domain,
		cfg.Nostr.PublicKey,
		cfg.LNURL.MaxSendableMsat,
		cfg.LNURL.MinSendableMsat,
		cfg.LNURL.CommentAllowed,
	)
}

func ProvideLNURLInvoiceHandler(cfg *config.Config, lndClient *lndrest.Client) server.LNURLInvoiceHandler {
	return server.NewLNURLInvoiceHandler(
		lndClient,
		cfg.Username,
		cfg.Nostr.PublicKey,
	)
}

func ProvideNostrHandler(cfg *config.Config) server.NostrHandler {
	return server.NewNostrHandler(cfg.Username, cfg.Nostr.PublicKey)
}

func main() {
	container := dig.New()

	if err := container.Provide(config.NewConfig); err != nil {
		panic(err)
	}

	if err := container.Provide(ProvideLNDClient); err != nil {
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

	if err := container.Invoke(func(router server.Router) error {
		return router.ListenAndServe(":8080")
	}); err != nil {
		panic(err)
	}
}
