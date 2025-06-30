package config

import (
	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
	"log/slog"
)

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		slog.Error("failed to parse config", "error", err)
		return nil, err
	}
	return cfg, nil
}

type Config struct {
	Server   ServerConfig
	LND      LNDConfig
	Username string `env:"USERNAME,notEmpty"`
	Nostr    NostrConfig
	LNURL    LNURLConfig
}

type ServerConfig struct {
	Host string `env:"SERVER_HOST,notEmpty"`
	Port string `env:"SERVER_PORT,notEmpty"`
}

type LNURLConfig struct {
	Domain          string `env:"DOMAIN,notEmpty"`
	MinSendableMsat int64  `env:"MIN_SENDABLE_MSAT" envDefault:"1000"`
	MaxSendableMsat int64  `env:"MAX_SENDABLE_MSAT" envDefault:"1000000000"`
	CommentAllowed  int64  `env:"COMMENT_ALLOWED" envDefault:"255"`
}

type LNDConfig struct {
	Host         string `env:"LND_HOST" envDefault:"http://localhost:8080"`
	MacaroonPath string `env:"LND_MACAROON_PATH" envDefault:"~/.lnd/data/chain/bitcoin/mainnet/admin.macaroon"`
}

type NostrConfig struct {
	PrivateKey string   `env:"NOSTR_PRIVATE_KEY,notEmpty"`
	PublicKey  string   `env:"NOSTR_PUBLIC_KEY,notEmpty"`
	Relays     []string `env:"NOSTR_RELAYS,notEmpty" envSeparator:","`
}
