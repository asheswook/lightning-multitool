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

// NewConfig 함수는 설정 파일을 읽고, 환경 변수를 적용한 후, 명령행 인자를 파싱하여 최종 설정을 반환합니다.
func NewConfig() (*Config, error) {
	// 1단계: 기본값으로 Config 구조체 초기화
	cfg := &Config{}
	parser := flags.NewParser(cfg, flags.Default)

	// 2단계: 설정 파일 경로를 먼저 찾기 위해 임시 파서 사용
	// IgnoreUnknown 옵션으로 다른 플래그는 무시하고 에러를 방지
	var pathOpts struct {
		ConfigFile string `short:"c" long:"config" description:"Path to config file" default:"lmt.conf" env:"LMT_CONFIG_FILE"`
	}
	pathParser := flags.NewParser(&pathOpts, flags.IgnoreUnknown)
	_, err := pathParser.ParseArgs(os.Args[1:])
	if err != nil {
		// 심각한 오류가 아니면 무시하고 진행
		slog.Debug("could not parse config path, using defaults", "error", err)
	}

	// 3단계: 찾은 경로로 INI 파일 파싱 (기본값을 덮어씀)
	iniParser := flags.NewIniParser(parser)
	if err := iniParser.ParseFile(pathOpts.ConfigFile); err != nil {
		// 설정 파일이 없는 경우는 에러가 아님 (명령행으로만 설정 가능)
		// 단, 기본 경로가 아닌 사용자가 명시적으로 지정한 파일이 없을 때는 경고를 표시할 수 있음
		if !os.IsNotExist(err) {
			slog.Warn("failed to parse config file", "path", pathOpts.ConfigFile, "error", err)
		}
	}

	// 4단계: 전체 명령행 파싱 (INI 값을 덮어씀)
	// 이 단계에서 최종적으로 모든 설정이 적용되고 유효성 검사가 이루어짐
	_, err = parser.ParseArgs(os.Args[1:])
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		slog.Error("failed to parse config flags", "error", err)
		return nil, err
	}

	return cfg, nil
}

type Config struct {
	ConfigFile string        `short:"c" long:"config" description:"Path to config file" default:"lmt.conf" env:"LMT_CONFIG_FILE"`
	General    GeneralConfig `group:"General" namespace:"general"`
	Server     ServerConfig  `group:"Server" namespace:"server"`
	API        APIConfig     `group:"API" namespace:"api"`
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

type APIConfig struct {
	Port string `long:"api_port" env:"API_PORT" description:"API port" default:"5051"`
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
	Enabled    bool     `long:"enable" env:"NOSTR_ENABLE"`
	PrivateKey string   `long:"privatekey" env:"NOSTR_PRIVATE_KEY" description:"Nostr private key (nsec format)"`
	PublicKey  string   `long:"publickey" env:"NOSTR_PUBLIC_KEY" description:"Nostr public key (npub format)"`
	Relays     []string `long:"relays" env:"NOSTR_RELAYS" env-delim:"," description:"Comma-separated list of Nostr relays" default:"wss://relay.damus.io,wss://relay.primal.net"`
}

type OksusuConfig struct {
	Enabled bool   `long:"enable" env:"OKSUSU_ENABLE" description:"Enable Oksusu integration"`
	Server  string `long:"server" env:"OKSUSU_SERVER" description:"Oksusu server" default:"oksu.su"`
	Token   string `long:"token" env:"OKSUSU_TOKEN" description:"Your Oksu Connect authentication token"`
}
