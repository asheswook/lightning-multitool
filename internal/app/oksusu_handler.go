package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"lmt/pkg/lndrest"
	nostrpkg "lmt/pkg/nostr"
	"lmt/pkg/oksusu" // The package we defined earlier

	"github.com/nbd-wtf/go-nostr"
)

type OksusuHandler struct {
	username       string
	host           string
	nostrPublicKey string

	maxSendable    int64
	minSendable    int64
	commentAllowed int64

	lndService *lndrest.Client
	zapMonitor ZapMonitor
}

// NewOksusuHandler creates a new OksusuHandler.
func NewOksusuHandler(username, host, nostrPublicKey string, maxSendable, minSendable, commentAllowed int64, lndService *lndrest.Client, zapMonitor ZapMonitor) OksusuHandler {
	return OksusuHandler{
		username:       username,
		host:           host,
		nostrPublicKey: nostrPublicKey,
		maxSendable:    maxSendable,
		minSendable:    minSendable,
		commentAllowed: commentAllowed,
		lndService:     lndService,
		zapMonitor:     zapMonitor,
	}
}

// OnLNURLPRequest handles the LNURL pay-request forwarded from the Oksusu server.
func (h OksusuHandler) OnLNURLPRequest(ctx context.Context, _ *oksusu.LNURLRequestPayload) (*oksusu.LNURLResponsePayload, error) {
	identifier := fmt.Sprintf("%s@%s", h.username, h.host)
	description := "Pay to " + identifier

	metadata := [][]string{
		{"text/plain", description},
		{"text/identifier", identifier},
	}
	encodedMetadata, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	callbackURL := fmt.Sprintf("https://%s/.well-known/lnurlp/%s/callback", h.host, h.username)

	var allowsNostr *bool
	allowsNostr = nil
	if h.nostrPublicKey != "" {
		*allowsNostr = true
	}

	return &oksusu.LNURLResponsePayload{
		Callback:        callbackURL,
		MaxSendable:     h.maxSendable,
		MinSendable:     h.minSendable,
		EncodedMetadata: string(encodedMetadata),
		CommentAllowed:  h.commentAllowed,
		Tag:             "payRequest",
		AllowsNostr:     allowsNostr,
		NostrPubkey:     h.nostrPublicKey,
	}, nil
}

// OnInvoiceRequest handles the invoice creation request forwarded from the Oksu server.
func (h OksusuHandler) OnInvoiceRequest(ctx context.Context, payload *oksusu.InvoiceRequestPayload) (*oksusu.InvoiceResponsePayload, error) {
	params := lndrest.CreateInvoiceParams{
		ValueMsat: payload.AmountMsat,
	}

	// Handle Nostr Zap request if present.
	var nostrEvent nostr.Event
	if payload.NostrZap != "" {
		if err := json.Unmarshal([]byte(payload.NostrZap), &nostrEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal nostr event: %w", err)
		}

		if _, err := nostrpkg.ParseZapRequest(nostrEvent, h.nostrPublicKey); err != nil {
			return nil, fmt.Errorf("invalid zap request: %w", err)
		}

		descriptionHash := sha256.Sum256([]byte(payload.NostrZap))
		params.DescriptionHash = descriptionHash[:]
		params.Expiry = 300 // 5 minutes for zap invoices
	}

	if payload.Comment != "" {
		params.Memo = payload.Comment
	}

	res, err := h.lndService.CreateInvoice(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// If it was a zap, start monitoring for payment to send a receipt.
	if payload.NostrZap != "" {
		go h.zapMonitor.MonitorAndSendZapReceipt(
			context.Background(), // Run in background
			res.RHash,
			nostrEvent,
			payload.NostrZap,
		)
	}

	return &oksusu.InvoiceResponsePayload{
		PR:     res.PaymentRequest,
		Routes: []interface{}{}, // Must be empty per LNURL spec
	}, nil
}
