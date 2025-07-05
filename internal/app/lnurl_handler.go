package app

import (
	"encoding/json"
	"fmt"
	"lmt/pkg/lnurl"
	"net/http"
)

type LNURLHandler struct {
	username       string
	domain         string
	maxSendable    int64
	minSendable    int64
	commentAllowed int64
	nostrPublicKey string
}

func NewLNURLHandler(username, domain, nostrPublicKey string, maxSendable, minSendable, commentAllowed int64) LNURLHandler {
	return LNURLHandler{
		username:       username,
		domain:         domain,
		maxSendable:    maxSendable,
		minSendable:    minSendable,
		commentAllowed: commentAllowed,
		nostrPublicKey: nostrPublicKey,
	}
}

func (h LNURLHandler) Handle(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("user")

	if username != h.username {
		response := lnurl.ErrorResponse{
			Status: "ERROR",
			Reason: "User not found",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	identifier := fmt.Sprintf("%s@%s", h.username, h.domain)
	description := "Send to " + identifier

	var metadata [][]string
	metadata = append(metadata, []string{"text/plain", description})
	metadata = append(metadata, []string{"text/identifier", identifier})

	j, _ := json.Marshal(metadata)

	response := lnurl.PayParamsWithNostr{
		PayParams: lnurl.PayParams{
			Response:        lnurl.Response{Status: "OK"},
			Callback:        fmt.Sprintf("https://%s/.well-known/lnurlp/%s/callback", h.domain, h.username),
			MaxSendable:     h.maxSendable,
			MinSendable:     h.minSendable,
			EncodedMetadata: string(j),
			CommentAllowed:  h.commentAllowed,
			Tag:             "payRequest",
		},
		AllowsNostr: true,
		NostrPubkey: h.nostrPublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
