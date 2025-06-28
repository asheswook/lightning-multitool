package server

import (
	"encoding/json"
	"fmt"
	"lmt/internal/config"
	"lmt/pkg/lnurl"
	"net/http"
)

func handleLNURLPay(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("user")

	if username != config.Cfg.Nostr.Username {
		response := lnurl.ErrorResponse{
			Status: "ERROR",
			Reason: "User not found",
		}

		_ = json.NewEncoder(w).Encode(response)
		return
	}

	identifier := fmt.Sprintf("%s@%s", config.Cfg.Nostr.Username, config.Cfg.LNURL.Domain)
	description := "Send to " + identifier

	var metadata [][]string
	mIdentifier := []string{"text/identifier", identifier}
	mDescription := []string{"text/plain", description}
	metadata = append(metadata, mIdentifier)
	metadata = append(metadata, mDescription)

	j, _ := json.Marshal(metadata)

	response := lnurl.PayParamsWithNostr{
		PayParams: lnurl.PayParams{
			Response:        lnurl.Response{Status: "OK"},
			Callback:        fmt.Sprintf("https://%s/.well-known/lnurlp/callback", config.Cfg.LNURL.Domain),
			MaxSendable:     config.Cfg.LNURL.MaxSendableMsat,
			MinSendable:     config.Cfg.LNURL.MinSendableMsat,
			EncodedMetadata: string(j),
			CommentAllowed:  config.Cfg.LNURL.CommentAllowed,
			Tag:             "payRequest",
		},
		AllowsNostr: true,
		NostrPubkey: config.Cfg.Nostr.PublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
