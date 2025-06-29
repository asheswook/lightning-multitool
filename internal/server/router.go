package server

import (
	"lmt/pkg/lndrest"
	"net/http"
)

func Router(client *lndrest.Client) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /.well-known/lnurlp/{user}", handleLNURLPay)
	router.HandleFunc("GET /.well-known/lnurlp/{user}/callback", func(w http.ResponseWriter, r *http.Request) {
		handleLNURLInvoice(client, w, r)
	})
	router.HandleFunc("GET /.well-known/nostr.json", handleNostrJSON)
	return router
}
