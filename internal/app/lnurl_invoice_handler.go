package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"github.com/asheswook/lightning-multitool/pkg/lndrest"
	"github.com/asheswook/lightning-multitool/pkg/lnurl"
	nostrpkg "github.com/asheswook/lightning-multitool/pkg/nostr"
	"github.com/nbd-wtf/go-nostr"
	"log/slog"
	"net/http"
	"strconv"
)

type LNURLInvoiceHandler struct {
	lndService     *lndrest.Client
	zapMonitor     ZapMonitor
	username       string
	nostrPublicKey string
}

func NewLNURLInvoiceHandler(lndService *lndrest.Client, zapMonitor ZapMonitor, username, nostrPublicKey string) LNURLInvoiceHandler {
	return LNURLInvoiceHandler{
		lndService:     lndService,
		zapMonitor:     zapMonitor,
		username:       username,
		nostrPublicKey: nostrPublicKey,
	}
}

// isNostrEnabled checks if Nostr functionality is enabled by checking if public key is set
func (h LNURLInvoiceHandler) isNostrEnabled() bool {
	return h.nostrPublicKey != ""
}

func (h LNURLInvoiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("user")
	if username != h.username {
		json.NewEncoder(w).Encode(lnurl.ErrorResponse{
			Status: "ERROR",
			Reason: "User not found",
		})
		return
	}

	amountStr := r.URL.Query().Get("amount")
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		json.NewEncoder(w).Encode(lnurl.ErrorResponse{
			Status: "ERROR",
			Reason: "Invalid amount parameter",
		})
		return
	}

	params := lndrest.CreateInvoiceParams{
		ValueMsat: amount,
	}

	var nostrEvent nostr.Event
	nostrParam := r.URL.Query().Get("nostr")
	if nostrParam != "" && h.isNostrEnabled() {
		if err := json.Unmarshal([]byte(nostrParam), &nostrEvent); err != nil {
			json.NewEncoder(w).Encode(lnurl.ErrorResponse{
				Status: "ERROR",
				Reason: "Failed to unmarshal nostr nostrEvent: " + err.Error(),
			})
			return
		}

		if _, err := nostrpkg.ParseZapRequest(nostrEvent, h.nostrPublicKey); err != nil {
			json.NewEncoder(w).Encode(lnurl.ErrorResponse{
				Status: "ERROR",
				Reason: "Invalid zap request: " + err.Error(),
			})
			return
		}

		// As per NIP-57, the description hash for a zap invoice is the sha256 hash of the zap request nostrEvent.
		descriptionHash := sha256.Sum256([]byte(nostrParam))
		params.DescriptionHash = descriptionHash[:]
		params.Expiry = 300 // 5 minutes
	} else if nostrParam != "" && !h.isNostrEnabled() {
		// Return error if Nostr parameter is provided but Nostr is disabled
		json.NewEncoder(w).Encode(lnurl.ErrorResponse{
			Status: "ERROR",
			Reason: "Nostr functionality is disabled",
		})
		return
	}

	commentParam := r.URL.Query().Get("comment")
	if commentParam != "" {
		params.Memo = commentParam
	}

	res, err := h.lndService.CreateInvoice(r.Context(), params)
	if err != nil {
		slog.Error("Failed to create invoice", "error", err)
		json.NewEncoder(w).Encode(lnurl.ErrorResponse{
			Status: "ERROR",
			Reason: "Failed to create invoice: " + err.Error(),
		})
		return
	}

	if nostrParam != "" && h.isNostrEnabled() {
		go h.zapMonitor.MonitorAndSendZapReceipt(
			context.Background(),
			res.RHash,
			nostrEvent,
			nostrParam,
		)
	}

	// According to LUD-06, the success response must be a JSON object
	// with a payment request (`pr`) and an empty `routes` array.
	response := lnurl.PayResponse{
		Response:   lnurl.Response{Status: "OK"},
		PR:         res.PaymentRequest,
		Routes:     []interface{}{},
		Disposable: false,
	}

	slog.Info("Responding with invoice", "amount", amount, "has_zap", nostrParam != "" && h.isNostrEnabled())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
