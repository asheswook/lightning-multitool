package server

import (
	"encoding/json"
	"lmt/pkg/nostr"
	"net/http"
)

type NostrHandler struct {
	username  string
	publicKey string
}

func NewNostrHandler(username, publicKey string) NostrHandler {
	return NostrHandler{
		username:  username,
		publicKey: publicKey,
	}
}

func (h NostrHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	nip5 := nostr.Nip5Data{
		Names: map[string]string{
			h.username: h.publicKey,
		},
	}

	json.NewEncoder(w).Encode(nip5)
}
