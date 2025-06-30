package server

import (
	"context"
	"encoding/hex"
	"lmt/internal/nostrutil"
	"lmt/pkg/lndrest"
	nostrspec "lmt/pkg/nostr"
	"log/slog"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

type ZapMonitor struct {
	lndService      *lndrest.Client
	nostrPrivateKey string
	nostrPublicKey  string
	relays          []string
}

func NewZapMonitor(lnd *lndrest.Client, pk, pubkey string, relays []string) ZapMonitor {
	return ZapMonitor{
		lndService:      lnd,
		nostrPrivateKey: pk,
		nostrPublicKey:  pubkey,
		relays:          relays,
	}
}

func (zm ZapMonitor) MonitorAndSendZapReceipt(
	ctx context.Context,
	paymentHash []byte,
	originalZapRequest nostr.Event,
	originalZapRequestRaw string,
) {
	paymentHashHex := hex.EncodeToString(paymentHash)
	logger := slog.With("payment_hash", paymentHashHex, "zap_request_id", originalZapRequest.ID)

	monitoringCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	invoiceChan, err := zm.lndService.SubscribeInvoices(monitoringCtx, lndrest.SubscribeInvoicesParams{})
	if err != nil {
		logger.Error("Failed to subscribe to invoices", "error", err)
		return
	}
	logger.Info("Started monitoring invoice payment for ZAP")

	for {
		select {
		case <-monitoringCtx.Done():
			logger.Warn("Stopped monitoring due to timeout or cancellation", "reason", monitoringCtx.Err())
			return
		case invoice, ok := <-invoiceChan:
			if !ok {
				logger.Warn("Invoice subscription channel closed unexpectedly")
				return
			}

			if hex.EncodeToString(invoice.RHash) == paymentHashHex && invoice.State == lndrest.InvoiceState_SETTLED {
				logger.Info("Invoice paid for ZAP", "amount_msat", invoice.AmtPaidMsat)
				zm.publishZapReceipt(invoice, originalZapRequest, originalZapRequestRaw)
				return // 임무 완료, 고루틴 종료
			}
		}
	}
}

func (zm ZapMonitor) publishZapReceipt(
	paidInvoice lndrest.Invoice,
	zapRequest nostr.Event,
	zapRequestRaw string,
) {
	preimageHex := hex.EncodeToString(paidInvoice.RPreimage)

	// pkg/nostr/zap.go에 정의된 NewZapReceipt 함수를 사용하여 Receipt를 생성합니다.
	receipt, err := nostrspec.NewZapReceipt(nostrspec.ZapReceiptParams{
		ZapRequest:       nostrspec.ZapRequest(zapRequest),
		Bolt11:           paidInvoice.PaymentRequest,
		Preimage:         preimageHex,
		RecipientPubkey:  zm.nostrPublicKey,
		RecipientPrivkey: zm.nostrPrivateKey,
	})
	if err != nil {
		slog.Error("Failed to create zap receipt", "error", err, "zap_request_id", zapRequest.ID)
		return
	}

	logger := slog.With("zap_receipt_id", receipt.ID, "zap_request_id", zapRequest.ID)
	logger.Info("Successfully created zap receipt, attempting to publish")

	nostrutil.PublishEvent(context.Background(), nostr.Event(receipt), zm.relays)
}
