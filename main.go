package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/fiatjaf/makeinvoice"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nbd-wtf/go-nostr"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/sjson"
)

// 기존 Config 구조체들 유지...
type Config struct {
	Server    ServerConfig    `json:"server"`
	Lightning LightningConfig `json:"lightning"`
	User      UserConfig      `json:"user"`
	Nostr     NostrConfig     `json:"nostr"`
	Relays    []string        `json:"relays"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type LightningConfig struct {
	Kind            string `json:"kind"`
	Host            string `json:"host"`
	Key             string `json:"key"`
	MinSendableMsat int64  `json:"min_sendable_msat"`
	MaxSendableMsat int64  `json:"max_sendable_msat"`
}

type UserConfig struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type NostrConfig struct {
	Username   string `json:"username"`
	Pubkey     string `json:"pubkey"`
	PrivateKey string `json:"private_key"` // Schnorr 서명용 개인키 추가
}

type Params struct {
	Kind        string `json:"kind"`
	Host        string `json:"host"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	MinSendable string `json:"min_sendable,omitempty"`
	MaxSendable string `json:"max_sendable,omitempty"`
}

// zap request 저장을 위한 구조체
type ZapRequest struct {
	Event     *nostr.Event `json:"event"`
	Amount    int64        `json:"amount"`
	Timestamp time.Time    `json:"timestamp"`
}

// 글로벌 변수들
var (
	config = Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "5000",
		},
		Lightning: LightningConfig{
			Kind:            "lnd",
			Host:            "https://localhost:8080",
			Key:             os.Getenv("LND_MACAROON"),
			MinSendableMsat: 1000,
			MaxSendableMsat: 1000000000,
		},
		User: UserConfig{
			Name:   "a",
			Domain: "pororo.ro",
		},
		Nostr: NostrConfig{
			Username:   "a",
			Pubkey:     os.Getenv("NOSTR_PUBKEY"),
			PrivateKey: os.Getenv("NOSTR_PRIVKEY"), // 실제 개인키로 교체 필요
		},
		Relays: []string{
			"wss://relay.damus.io",
			"wss://nostr.mom",
			"wss://nos.lol",
			"wss://relay.primal.net",
			"wss://purplepag.es",
			"wss://nostr.wine",
			"wss://relay.nostr.band",
		},
	}

	// zap request들을 임시 저장
	zapRequests = make(map[string]*ZapRequest)
	zapMutex    = sync.RWMutex{}
)

func main() {
	initializeServer()
	router := setupRoutes()
	startServer(router)
}

func initializeServer() {
	makeinvoice.Client = &http.Client{Timeout: 25 * time.Second}
	log.Info().Str("module", "server").Msg("Server initialized")
}

func convertPrivateKey(key string) (string, error) {
	if strings.HasPrefix(key, "nsec1") {
		// bech32 직접 디코딩
		_, data, err := bech32.Decode(key)
		if err != nil {
			return "", fmt.Errorf("failed to decode bech32: %w", err)
		}

		// 5bit to 8bit 변환
		converted, err := bech32.ConvertBits(data, 5, 8, false)
		if err != nil {
			return "", fmt.Errorf("failed to convert bits: %w", err)
		}

		return hex.EncodeToString(converted), nil
	}

	// 이미 hex 형식이면 그대로 반환
	if len(key) == 64 {
		if _, err := hex.DecodeString(key); err != nil {
			return "", fmt.Errorf("invalid hex private key: %w", err)
		}
		return key, nil
	}

	return "", fmt.Errorf("unsupported private key format")
}

func setupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.Path("/.well-known/lnurlp/{user}").
		Methods("GET").
		HandlerFunc(handleLNURLPay)

	router.Path("/.well-known/nostr.json").
		Methods("GET").
		HandlerFunc(handleNostrJSON)

	return router
}

func startServer(router *mux.Router) {
	addr := config.Server.Host + ":" + config.Server.Port
	logServerInfo(addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

func logServerInfo(addr string) {
	log.Info().
		Str("module", "server").
		Str("address", addr).
		Msg("Starting LNURL server with zap support")

	log.Info().Msgf("LNURL endpoint: http://%s/.well-known/lnurlp/%s", addr, config.User.Name)
	log.Info().Msgf("Lightning Address: %s@%s", config.User.Name, config.User.Domain)
}

func handleLNURLPay(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]

	if !isValidUser(username) {
		respondWithError(w, "User not found")
		return
	}

	amountStr := r.URL.Query().Get("amount")
	if amountStr == "" {
		handlePayParamsRequest(w, username)
	} else {
		handleInvoiceRequest(w, username, amountStr, r)
	}
}

func isValidUser(username string) bool {
	return username == config.User.Name
}

func handlePayParamsRequest(w http.ResponseWriter, username string) {
	params := getUserParams()
	metadata := generateMetadata(params)
	callbackURL := fmt.Sprintf("https://%s/.well-known/lnurlp/%s", config.User.Domain, username)

	response := createPayParamsResponse(callbackURL, metadata)

	log.Info().Str("module", "lnurl-handler").Msg("Responding with LNURLPayParams")
	json.NewEncoder(w).Encode(response)
}

func createPayParamsResponse(callbackURL, metadata string) interface{} {
	baseResponse := lnurl.LNURLPayParams{
		LNURLResponse:   lnurl.LNURLResponse{Status: "OK"},
		Callback:        callbackURL,
		MinSendable:     config.Lightning.MinSendableMsat,
		MaxSendable:     config.Lightning.MaxSendableMsat,
		EncodedMetadata: metadata,
		CommentAllowed:  0,
		Tag:             "payRequest",
	}

	return struct {
		lnurl.LNURLPayParams
		AllowsNostr bool   `json:"allowsNostr"`
		NostrPubkey string `json:"nostrPubkey"`
	}{
		LNURLPayParams: baseResponse,
		AllowsNostr:    true,
		NostrPubkey:    config.Nostr.Pubkey,
	}
}

// zap 요청 처리가 포함된 인보이스 핸들러
func handleInvoiceRequest(w http.ResponseWriter, username, amountStr string, r *http.Request) {
	amount, err := parseAndValidateAmount(amountStr)
	if err != nil {
		respondWithError(w, err.Error())
		return
	}

	// zap request 처리
	nostrParam := r.URL.Query().Get("nostr")
	var zapReq *ZapRequest
	if nostrParam != "" {
		zapReq, err = processZapRequest(nostrParam, amount)
		if err != nil {
			log.Error().Err(err).Msg("Invalid zap request")
			respondWithError(w, "Invalid zap request: "+err.Error())
			return
		}
	}

	params := getUserParams()
	bolt11, err := createInvoice(params, int(amount), zapReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create invoice")
		respondWithError(w, "Failed to create invoice: "+err.Error())
		return
	}

	// zap request가 있으면 저장
	if zapReq != nil {
		zapMutex.Lock()
		zapRequests[bolt11] = zapReq
		zapMutex.Unlock()

		// 결제 확인을 위한 고루틴 시작
		go monitorPayment(bolt11, zapReq)
	}

	response := createInvoiceResponse(bolt11)

	log.Info().
		Str("module", "lnurl-handler").
		Int64("amount_msat", amount).
		Bool("has_zap", zapReq != nil).
		Msg("Responding with invoice")

	json.NewEncoder(w).Encode(response)
}

// zap request 검증 및 처리[3]
func processZapRequest(nostrParam string, amount int64) (*ZapRequest, error) {
	decoded, err := url.QueryUnescape(nostrParam)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nostr parameter: %w", err)
	}

	var event nostr.Event
	if err := json.Unmarshal([]byte(decoded), &event); err != nil {
		return nil, fmt.Errorf("failed to parse zap request: %w", err)
	}

	// NIP-57 검증[6]
	if event.Kind != 9734 {
		return nil, fmt.Errorf("invalid kind, expected 9734, got %d", event.Kind)
	}

	// 서명 검증
	if ok, err := event.CheckSignature(); !ok {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	// 태그 검증
	var hasP, hasAmount bool
	var pTag, amountTag string

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		switch tag[0] {
		case "p":
			if hasP {
				return nil, fmt.Errorf("multiple p tags not allowed")
			}
			hasP = true
			pTag = tag[1]
		case "amount":
			hasAmount = true
			amountTag = tag[1]
		}
	}

	if !hasP {
		return nil, fmt.Errorf("missing p tag")
	}

	if pTag != config.Nostr.Pubkey {
		return nil, fmt.Errorf("p tag does not match recipient pubkey")
	}

	if hasAmount {
		reqAmount, err := strconv.ParseInt(amountTag, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount tag: %w", err)
		}
		if reqAmount != amount {
			return nil, fmt.Errorf("amount mismatch: requested %d, got %d", reqAmount, amount)
		}
	}

	return &ZapRequest{
		Event:     &event,
		Amount:    amount,
		Timestamp: time.Now(),
	}, nil
}

// 인보이스 생성 (zap request 포함)
func createInvoice(params *Params, msat int, zapReq *ZapRequest) (string, error) {
	backend, err := createBackendParams(params)
	if err != nil {
		return "", err
	}

	var description string
	var useDescriptionHash bool

	if zapReq != nil {
		// zap request를 description으로 사용
		zapReqBytes, _ := json.Marshal(zapReq.Event)
		description = string(zapReqBytes)
		useDescriptionHash = true
	} else {
		description = generateMetadata(params)
		useDescriptionHash = true
	}

	invoiceParams := makeinvoice.Params{
		Msatoshi:           int64(msat),
		Backend:            backend,
		Label:              generateInvoiceLabel(params),
		UseDescriptionHash: useDescriptionHash,
		Description:        description,
	}

	log.Info().
		Str("module", "invoice-generator").
		Bool("is_zap", zapReq != nil).
		Int64("msatoshi", invoiceParams.Msatoshi).
		Msg("Generating invoice")

	bolt11, err := makeinvoice.MakeInvoice(invoiceParams)
	if err != nil {
		return "", fmt.Errorf("makeinvoice library call failed: %w", err)
	}

	return bolt11, nil
}

// 결제 모니터링 및 zap receipt 생성
func monitorPayment(bolt11 string, zapReq *ZapRequest) {
	// 실제 구현에서는 LND 인보이스 상태를 구독해야 함
	// 여기서는 예시로 5초 후 결제됐다고 가정
	time.Sleep(5 * time.Second)

	log.Info().Str("bolt11", bolt11[:20]+"...").Msg("Payment detected, creating zap receipt")

	if err := createAndPublishZapReceipt(bolt11, zapReq); err != nil {
		log.Error().Err(err).Msg("Failed to create/publish zap receipt")
	}

	// 정리
	zapMutex.Lock()
	delete(zapRequests, bolt11)
	zapMutex.Unlock()
}

// zap receipt 생성 및 발행[3][6]
func createAndPublishZapReceipt(bolt11 string, zapReq *ZapRequest) error {
	now := time.Now()

	// 태그 구성
	tags := nostr.Tags{
		{"p", getRecipientFromZapRequest(zapReq.Event)},
		{"bolt11", bolt11},
		{"description", string(mustMarshal(zapReq.Event))},
	}

	// 원본 zap request의 태그들 복사
	for _, tag := range zapReq.Event.Tags {
		switch tag[0] {
		case "e", "a": // 이벤트/좌표 태그 복사
			tags = append(tags, tag)
		case "P": // 송신자 pubkey 복사
			tags = append(tags, tag)
		case "relays": // 릴레이 태그 복사
			tags = append(tags, tag)
		}
	}

	// relays 태그가 없으면 기본 릴레이 추가
	hasRelays := false
	for _, tag := range tags {
		if tag[0] == "relays" {
			hasRelays = true
			break
		}
	}
	if !hasRelays {
		relayTag := []string{"relays"}
		relayTag = append(relayTag, config.Relays...)
		tags = append(tags, relayTag)
	}

	// zap receipt 이벤트 생성
	receipt := nostr.Event{
		PubKey:    config.Nostr.Pubkey,
		CreatedAt: nostr.Timestamp(now.Unix()),
		Kind:      9735,
		Tags:      tags,
		Content:   "",
	}

	privateKeyHex, err := convertPrivateKey(config.Nostr.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to convert private key: %w", err)
	}

	// 서명
	if err := receipt.Sign(privateKeyHex); err != nil {
		return fmt.Errorf("failed to sign zap receipt: %w", err)
	}

	// 릴레이에 발행
	return publishToRelays(&receipt, config.Relays)
}

// 릴레이에 이벤트 발행[11]
func publishToRelays(event *nostr.Event, relays []string) error {
	var wg sync.WaitGroup
	var errors []error
	var mu sync.Mutex

	for _, relayURL := range relays {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := publishToRelay(event, url); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("relay %s: %w", url, err))
				mu.Unlock()
			} else {
				log.Info().Str("relay", url).Msg("Successfully published zap receipt")
			}
		}(relayURL)
	}

	wg.Wait()

	if len(errors) == len(relays) {
		return fmt.Errorf("failed to publish to all relays: %v", errors)
	}

	return nil
}

func publishToRelay(event *nostr.Event, relayURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, relayURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// EVENT 메시지 전송
	eventMsg := []interface{}{"EVENT", event}
	if err := conn.WriteJSON(eventMsg); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	// OK 응답 대기
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var response []interface{}
	if err := conn.ReadJSON(&response); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if len(response) >= 3 && response[0] == "OK" {
		if success, ok := response[2].(bool); ok && success {
			return nil
		}
		if len(response) >= 4 {
			return fmt.Errorf("relay rejected: %v", response[3])
		}
	}

	return fmt.Errorf("unexpected response: %v", response)
}

// 헬퍼 함수들
func getRecipientFromZapRequest(zapReq *nostr.Event) string {
	for _, tag := range zapReq.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			return tag[1]
		}
	}
	return config.Nostr.Pubkey // 기본값
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// 기존 함수들 유지...
func parseAndValidateAmount(amountStr string) (int64, error) {
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("amount is not a valid integer")
	}

	if amount < config.Lightning.MinSendableMsat || amount > config.Lightning.MaxSendableMsat {
		return 0, fmt.Errorf("amount must be between %d and %d msat",
			config.Lightning.MinSendableMsat, config.Lightning.MaxSendableMsat)
	}

	return amount, nil
}

func createInvoiceResponse(bolt11 string) lnurl.LNURLPayValues {
	return lnurl.LNURLPayValues{
		LNURLResponse: lnurl.LNURLResponse{Status: "OK"},
		PR:            bolt11,
		Routes:        make([]interface{}, 0),
		Disposable:    lnurl.FALSE,
		SuccessAction: &lnurl.SuccessAction{
			Tag:     "message",
			Message: "Pororo is now happy 😊",
		},
	}
}

func createBackendParams(params *Params) (makeinvoice.BackendParams, error) {
	switch params.Kind {
	case "lnd":
		return makeinvoice.LNDParams{
			Host:     params.Host,
			Macaroon: params.Key,
			Private:  true,
		}, nil
	case "lnbits":
		return makeinvoice.LNBitsParams{
			Host: params.Host,
			Key:  params.Key,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Lightning backend kind: %s", params.Kind)
	}
}

func generateInvoiceLabel(params *Params) string {
	return fmt.Sprintf("%s/%s/%s",
		params.Domain,
		params.Name,
		strconv.FormatInt(time.Now().UnixNano(), 16))
}

func handleNostrJSON(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Names map[string]string `json:"names"`
	}{
		Names: map[string]string{
			config.Nostr.Username: config.Nostr.Pubkey,
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewEncoder(w).Encode(response)
}

func getUserParams() *Params {
	return &Params{
		Kind:        config.Lightning.Kind,
		Host:        config.Lightning.Host,
		Key:         config.Lightning.Key,
		Name:        config.User.Name,
		Domain:      config.User.Domain,
		MinSendable: strconv.FormatInt(config.Lightning.MinSendableMsat, 10),
		MaxSendable: strconv.FormatInt(config.Lightning.MaxSendableMsat, 10),
	}
}

func generateMetadata(params *Params) string {
	identifier := params.Name + "@" + params.Domain
	description := "Send to " + identifier

	metadata := "[]"
	metadata, _ = sjson.Set(metadata, "0.0", "text/identifier")
	metadata, _ = sjson.Set(metadata, "0.1", identifier)
	metadata, _ = sjson.Set(metadata, "1.0", "text/plain")
	metadata, _ = sjson.Set(metadata, "1.1", description)

	return metadata
}

func respondWithError(w http.ResponseWriter, message string) {
	log.Warn().Str("error", message).Msg("Sending error response")
	json.NewEncoder(w).Encode(lnurl.ErrorResponse(message))
}
