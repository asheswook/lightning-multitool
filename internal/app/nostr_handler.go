package app

import (
	"encoding/json"
	"github.com/asheswook/lightning-multitool/pkg/nostr"
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

// isNostrEnabled checks if Nostr functionality is enabled by checking if public key is set
func (h NostrHandler) isNostrEnabled() bool {
	return h.publicKey != ""
}

func (h NostrHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !h.isNostrEnabled() {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Nostr functionality is disabled",
		})
		return
	}

	nip5 := nostr.Nip5Data{
		Names: map[string]string{
			h.username: h.publicKey,
		},
	}

	json.NewEncoder(w).Encode(nip5)
}
