package config

import (
	"github.com/jessevdk/go-flags"
	"log/slog"
	"os"
)

// pathOptions is a helper struct to only parse the config file path
// from command-line arguments or environment variables.
type pathOptions struct {
	ConfigFile string `short:"c" long:"config" description:"Path to config file" default:"lmt.conf" env:"LMT_CONFIG_FILE"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	parser := flags.NewParser(cfg, flags.Default)

	// 1단계: config 파일 경로만 먼저 파싱 (기존 방식 유지)
	pathOpts := &pathOptions{}
	_, err := flags.NewParser(pathOpts, flags.IgnoreUnknown).ParseArgs(os.Args[1:])
	if err != nil {
		slog.Error("failed to pre-parse for config file path", "error", err)
		return nil, err
	}

	// 2단계: INI 파일을 먼저 로드 (required 값들을 채움)
	if pathOpts.ConfigFile != "" {
		iniParser := flags.NewIniParser(parser)
		err := iniParser.ParseFile(pathOpts.ConfigFile)
		if err != nil {
			if os.IsNotExist(err) && pathOpts.ConfigFile == "lmt.conf" {
				// 기본 파일이 없는 것은 괜찮음
			} else {
				slog.Error("failed to parse config file", "path", pathOpts.ConfigFile, "error", err)
				return nil, err
			}
		}
	}

	// 3단계: 명령행 파싱 (INI 값을 덮어씀)
	_, err = parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		slog.Error("failed to parse config", "error", err)
		return nil, err
	}

	return cfg, nil
}

type Config struct {
	ConfigFile string        `short:"c" long:"config" description:"Path to config file" default:"lmt.conf" env:"LMT_CONFIG_FILE"`
	General    GeneralConfig `group:"General" namespace:"general"`
	Server     ServerConfig  `group:"Server" namespace:"server"`
	LND        LNDConfig     `group:"LND" namespace:"lnd"`
	Nostr      NostrConfig   `group:"Nostr" namespace:"nostr"`
	LNURL      LNURLConfig   `group:"LNURL" namespace:"lnurl"`
	Oksusu     OksusuConfig  `group:"Oksusu" namespace:"oksusu"`
}

type GeneralConfig struct {
	Username string `long:"username" env:"USERNAME" description:"Username for the Lightning Address" required:"true"`
}

type ServerConfig struct {
	Host string `long:"host" env:"SERVER_HOST" description:"Server host"`
	Port string `long:"port" env:"SERVER_PORT" description:"Server port"`
}

type LNURLConfig struct {
	Domain          string `long:"domain" env:"DOMAIN" description:"Domain for the LNURL" required:"true"`
	MinSendableMsat int64  `long:"min-sendable" env:"MIN_SENDABLE_MSAT" description:"Minimum sendable amount in msats" default:"1000"`
	MaxSendableMsat int64  `long:"max-sendable" env:"MAX_SENDABLE_MSAT" description:"Maximum sendable amount in msats" default:"1000000000"`
	CommentAllowed  int64  `long:"comment-allowed" env:"COMMENT_ALLOWED" description:"Maximum comment length" default:"255"`
}

type LNDConfig struct {
	Host         string `long:"host" env:"LND_HOST" description:"LND REST host" default:"localhost:8080"`
	MacaroonPath string `long:"macaroonpath" env:"LND_MACAROON_PATH" description:"Path to LND admin.macaroon" default:"~/.lnd/data/chain/bitcoin/mainnet/admin.macaroon"`
}

type NostrConfig struct {
	PrivateKey string   `long:"privatekey" env:"NOSTR_PRIVATE_KEY" description:"Nostr private key (nsec format)"`
	PublicKey  string   `long:"publickey" env:"NOSTR_PUBLIC_KEY" description:"Nostr public key (npub format)"`
	Relays     []string `long:"relays" env:"NOSTR_RELAYS" env-delim:"," description:"Comma-separated list of Nostr relays" default:"wss://relay.damus.io,wss://relay.primal.net"`
}

type OksusuConfig struct {
	Enabled bool   `long:"enabled" env:"OKSUSU_ENABLED" description:"Enable Oksusu integration"`
	Server  string `long:"server" env:"OKSUSU_SERVER" description:"Oksusu server" default:"oksu.su"`
	Token   string `long:"token" env:"OKSUSU_TOKEN" description:"Your Oksu Connect authentication token"`
}
