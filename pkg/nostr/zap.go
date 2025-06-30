package nostr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/nbd-wtf/go-nostr"
)

func ParseZapRequest(event nostr.Event, recipientPubkey string) (ZapRequest, error) {
	if event.Kind != 9734 {
		return ZapRequest{}, fmt.Errorf("invalid kind, expected 9734, got %d", event.Kind)
	}

	if ok, err := event.CheckSignature(); !ok {
		return ZapRequest{}, fmt.Errorf("invalid signature: %w", err)
	}

	p := event.Tags.Find("p")
	if p == nil || len(p) < 2 {
		return ZapRequest{}, fmt.Errorf("missing p tag or invalid format")
	}

	if p[1] != recipientPubkey {
		return ZapRequest{}, fmt.Errorf("p tag pubkey '%s' does not match recipient pubkey '%s'", p[1], recipientPubkey)
	}

	return ZapRequest(event), nil
}

type ZapRequest nostr.Event

func (z ZapRequest) Event() nostr.Event {
	return nostr.Event(z)
}

type ZapReceiptParams struct {
	ZapRequest       ZapRequest
	Bolt11           string // Paid Invoice
	Preimage         string // Preimage (hex-encoded)
	RecipientPubkey  string // Public key of recipient
	RecipientPrivkey string // Private key to sign
}

func NewZapReceipt(params ZapReceiptParams) (ZapReceipt, error) {
	zapRequestJSON, err := json.Marshal(params.ZapRequest)
	if err != nil {
		return ZapReceipt{}, fmt.Errorf("failed to marshal original zap request: %w", err)
	}

	tags := nostr.Tags{
		{"p", params.ZapRequest.PubKey},
		{"bolt11", params.Bolt11},
		{"description", string(zapRequestJSON)},
		{"preimage", params.Preimage},
	}

	if eTag := params.ZapRequest.Tags.Find("e"); eTag != nil {
		tags = append(tags, eTag)
	}

	event := nostr.Event{
		Kind:      9735,
		Tags:      tags,
		PubKey:    params.RecipientPubkey,
		CreatedAt: nostr.Now(),
		Content:   "", // intended
	}

	if err = event.Sign(params.RecipientPrivkey); err != nil {
		return ZapReceipt{}, fmt.Errorf("failed to sign zap event: %w", err)
	}

	receipt := ZapReceipt(event)
	if err = receipt.Validate(params.ZapRequest.PubKey); err != nil {
		return ZapReceipt{}, fmt.Errorf("failed to validate zap receipt: %w", err)
	}

	return receipt, nil
}

type ZapReceipt nostr.Event

func (z ZapReceipt) Event() nostr.Event {
	return nostr.Event(z)
}

func (z ZapReceipt) Validate(zappedUserPubkey string) error {
	if z.Kind != 9735 {
		return fmt.Errorf("invalid kind: expected 9735, got %d", z.Kind)
	}

	if ok, _ := z.Event().CheckSignature(); !ok {
		return fmt.Errorf("invalid signature on receipt event")
	}

	p := z.Tags.Find("p")
	bolt11 := z.Tags.Find("bolt11")
	description := z.Tags.Find("description")
	preimageTag := z.Tags.Find("preimage")

	if p == nil || bolt11 == nil || description == nil || preimageTag == nil {
		return fmt.Errorf("missing one or more required tags (p, bolt11, preimage, description)")
	}

	if p[1] != zappedUserPubkey {
		return fmt.Errorf("p tag pubkey '%s' does not match zapped user '%s'", p[1], zappedUserPubkey)
	}

	preimage, err := hex.DecodeString(preimageTag[1])
	if err != nil || len(preimage) != 32 {
		return fmt.Errorf("invalid preimage format: %w", err)
	}
	hash := sha256.Sum256(preimage)

	invoice, err := zpay32.Decode(bolt11[1], &chaincfg.MainNetParams)
	if err != nil {
		return fmt.Errorf("failed to decode bolt11 invoice: %w", err)
	}

	if *invoice.PaymentHash != hash {
		return fmt.Errorf("preimage hash does not match invoice payment_hash")
	}

	return nil
}
