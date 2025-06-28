package server

import (
	"encoding/json"
	"lmt/internal/config"
	"lmt/pkg/nostr"
	"net/http"
)

func handleNostrJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	nip5 := nostr.Nip5Data{
		Names: map[string]string{
			config.Cfg.Nostr.Username: config.Cfg.Nostr.PublicKey,
		},
	}

	_ = json.NewEncoder(w).Encode(nip5)
}
