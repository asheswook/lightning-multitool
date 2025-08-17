package oksusu

import "encoding/json"

// MessageType is a type of WebSocket message.
type MessageType string

const (
	// Client to Server (C2S)
	C2SAuth            MessageType = "c2s_auth"
	C2SLNURLPResponse  MessageType = "c2s_lnurlp_response"
	C2SInvoiceResponse MessageType = "c2s_invoice_response"
	C2SError           MessageType = "c2s_error"

	// Server to Client (S2C)
	S2CAuthOK         MessageType = "s2c_auth_ok"
	S2CAuthFail       MessageType = "s2c_auth_fail"
	S2CLNURLPRequest  MessageType = "s2c_lnurlp_request"
	S2CInvoiceRequest MessageType = "s2c_invoice_request"
	S2CError          MessageType = "s2c_error"
)

type Message struct {
	ID      string          `json:"id"`
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type AuthRequestPayload struct {
	Token string `json:"token"`
}

type AuthResponsePayload struct {
	Username string `json:"username"`
	Message  string `json:"message,omitempty"`
}

type LNURLRequestPayload struct {
	// Empty
}

type LNURLResponsePayload struct {
	Callback        string `json:"callback"`
	MaxSendable     int64  `json:"maxSendable"`
	MinSendable     int64  `json:"minSendable"`
	EncodedMetadata string `json:"metadata"`
	CommentAllowed  int64  `json:"commentAllowed"`
	Tag             string `json:"tag"`
	AllowsNostr     *bool  `json:"allowsNostr,omitempty"` // 포인터를 사용해 false 값도 생략 가능하도록 함
	NostrPubkey     string `json:"nostrPubkey,omitempty"`
}

type InvoiceRequestPayload struct {
	AmountMsat int64  `json:"amount_msat"`
	Comment    string `json:"comment,omitempty"`
	NostrZap   string `json:"nostr_zap,omitempty"` // URL-decoded nostr event JSON string
}

type InvoiceResponsePayload struct {
	PR            string                `json:"pr"`
	Routes        []interface{}         `json:"routes"` // 항상 비어있어야 함
	SuccessAction *SuccessActionPayload `json:"successAction,omitempty"`
}

type SuccessActionPayload struct {
	Tag     string `json:"tag"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}
