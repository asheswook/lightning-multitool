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
	// First, parse the command line just for the config file path.
	pathOpts := &pathOptions{}
	// Use a temporary parser with IgnoreUnknown to avoid errors about other flags.
	// We parse os.Args[1:] to avoid the program name.
	_, err := flags.NewParser(pathOpts, flags.IgnoreUnknown).ParseArgs(os.Args[1:])
	if err != nil {
		slog.Error("failed to pre-parse for config file path", "error", err)
		return nil, err
	}

	// Now, create the main config struct and its parser.
	cfg := &Config{}
	parser := flags.NewParser(cfg, flags.Default)

	// Load settings from the config file found in the first pass.
	// The default value in pathOptions ensures we try "lmt.conf" if nothing is specified.
	if pathOpts.ConfigFile != "" {
		iniParser := flags.NewIniParser(parser)
		err := iniParser.ParseFile(pathOpts.ConfigFile)
		// If the file doesn't exist, we only ignore the error if it's the default path.
		// If the user explicitly provided a path, it must exist.
		if err != nil {
			if os.IsNotExist(err) && pathOpts.ConfigFile == "lmt.conf" {
				// Default config file doesn't exist, which is fine.
			} else {
				slog.Error("failed to parse config file", "path", pathOpts.ConfigFile, "error", err)
				return nil, err
			}
		}
	}

	// Finally, parse the command line arguments again. This time, the parser
	// knows about all the flags. It will apply overrides from the command line
	// and environment variables, and then check for all required flags.
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
	ConfigFile string       `short:"c" long:"config" description:"Path to config file" default:"lmt.conf" env:"LMT_CONFIG_FILE"`
	Server     ServerConfig `group:"Server Options"`
	LND        LNDConfig    `group:"LND Options"`
	Username   string       `long:"username" env:"USERNAME" description:"Username for the Lightning Address" required:"true"`
	Nostr      NostrConfig  `group:"Nostr Options"`
	LNURL      LNURLConfig  `group:"LNURL Options"`
}

type ServerConfig struct {
	Host string `long:"server.host" env:"SERVER_HOST" description:"Server host" required:"true"`
	Port string `long:"server.port" env:"SERVER_PORT" description:"Server port" required:"true"`
}

type LNURLConfig struct {
	Domain          string `long:"lnurl.domain" env:"DOMAIN" description:"Domain for the LNURL" required:"true"`
	MinSendableMsat int64  `long:"lnurl.min-sendable" env:"MIN_SENDABLE_MSAT" description:"Minimum sendable amount in msats" default:"1000"`
	MaxSendableMsat int64  `long:"lnurl.max-sendable" env:"MAX_SENDABLE_MSAT" description:"Maximum sendable amount in msats" default:"1000000000"`
	CommentAllowed  int64  `long:"lnurl.comment-allowed" env:"COMMENT_ALLOWED" description:"Maximum comment length" default:"255"`
}

type LNDConfig struct {
	Host         string `long:"lnd.host" env:"LND_HOST" description:"LND REST host" default:"localhost:8080"`
	MacaroonPath string `long:"lnd.macaroonpath" env:"LND_MACAROON_PATH" description:"Path to LND admin.macaroon" default:"~/.lnd/data/chain/bitcoin/mainnet/admin.macaroon"`
}

type NostrConfig struct {
	PrivateKey string   `long:"nostr.privatekey" env:"NOSTR_PRIVATE_KEY" description:"Nostr private key (nsec format)" required:"true"`
	PublicKey  string   `long:"nostr.publickey" env:"NOSTR_PUBLIC_KEY" description:"Nostr public key (npub format)" required:"true"`
	Relays     []string `long:"nostr.relays" env:"NOSTR_RELAYS" env-delim:"," description:"Comma-separated list of Nostr relays" default:"wss://relay.damus.io,wss://nostr-pub.wellorder.net"`
}
