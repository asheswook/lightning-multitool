package main

import (
	"encoding/hex"
	"fmt"
	"github.com/nbd-wtf/go-nostr/nip19"
	"go.uber.org/dig"
	"lmt/internal/app"
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
		cfg.Username,
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
		cfg.Username,
		vpub.(string),
	)
}

func ProvideNostrHandler(cfg *config.Config) app.NostrHandler {
	_, vpub, err := nip19.Decode(cfg.Nostr.PublicKey)
	if err != nil {
		panic(err)
	}

	return app.NewNostrHandler(cfg.Username, vpub.(string))
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

	if err := container.Invoke(func(cfg *config.Config, router server.Router) error {
		return router.ListenAndServe(cfg.Server.Host + ":" + cfg.Server.Port)
	}); err != nil {
		panic(err)
	}
}
