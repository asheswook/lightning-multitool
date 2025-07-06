package nostr

import (
	"fmt"
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
	ZapRequestRaw    string
	Bolt11           string // Paid Invoice
	Preimage         string // Preimage (hex-encoded)
	RecipientPubkey  string // Public key of recipient
	RecipientPrivkey string // Private key to sign
}

func NewZapReceipt(params ZapReceiptParams) (ZapReceipt, error) {
	tags := nostr.Tags{
		{"p", params.RecipientPubkey},
		{"bolt11", params.Bolt11},
		{"description", params.ZapRequestRaw},
	}

	if params.Preimage != "" {
		tags = append(tags, nostr.Tag{"preimage", params.Preimage})
	}

	if eTag := params.ZapRequest.Tags.Find("e"); eTag != nil {
		tags = append(tags, eTag)
	}

	if relaysTag := params.ZapRequest.Tags.Find("relays"); relaysTag != nil {
		tags = append(tags, relaysTag)
	}

	event := nostr.Event{
		Kind:      9735,
		Tags:      tags,
		PubKey:    params.RecipientPubkey,
		CreatedAt: nostr.Now(),
		Content:   "", // intended
	}

	if err := event.Sign(params.RecipientPrivkey); err != nil {
		return ZapReceipt{}, fmt.Errorf("failed to sign zap event: %w", err)
	}

	receipt := ZapReceipt(event)
	if err := receipt.Validate(); err != nil {
		return ZapReceipt{}, fmt.Errorf("failed to validate zap receipt: %w", err)
	}

	return receipt, nil
}

type ZapReceipt nostr.Event

func (z ZapReceipt) Event() nostr.Event {
	return nostr.Event(z)
}

func (z ZapReceipt) Validate() error {
	if z.Kind != 9735 {
		return fmt.Errorf("invalid kind: expected 9735, got %d", z.Kind)
	}

	if ok, _ := z.Event().CheckSignature(); !ok {
		return fmt.Errorf("invalid signature on receipt event")
	}

	p := z.Tags.Find("p")
	bolt11 := z.Tags.Find("bolt11")
	description := z.Tags.Find("description")

	if p == nil || bolt11 == nil || description == nil {
		return fmt.Errorf("missing one or more required tags (p, bolt11, preimage, description)")
	}

	return nil
}
