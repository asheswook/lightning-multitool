package server

import (
	"github.com/asheswook/lightning-multitool/internal/app"
	"log/slog"
	"net/http"
)

type Router struct {
	lnurlInvoiceHandler app.LNURLInvoiceHandler
	lnurlHandler        app.LNURLHandler
	nostrHandler        app.NostrHandler
}

func NewRouter(lnurlInvoiceHandler app.LNURLInvoiceHandler, lnurlHandler app.LNURLHandler, nostrHandler app.NostrHandler) Router {
	return Router{
		lnurlInvoiceHandler: lnurlInvoiceHandler,
		lnurlHandler:        lnurlHandler,
		nostrHandler:        nostrHandler,
	}
}

func (r Router) ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/lnurlp/{user}", r.lnurlHandler.Handle)
	mux.HandleFunc("/.well-known/nostr.json", r.nostrHandler.Handle)
	mux.HandleFunc("/.well-known/lnurlp/{user}/callback", r.lnurlInvoiceHandler.Handle)
	return mux
}

func (r Router) ListenAndServe(addr string) error {
	slog.Info("Listening on", "addr", addr)
	return http.ListenAndServe(addr, r.ServeMux())
}
