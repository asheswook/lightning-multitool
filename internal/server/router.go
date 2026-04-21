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

// withCORS adds permissive CORS headers required by LUD-01/LUD-16 and NIP-05
// so browser-based wallets can fetch these endpoints cross-origin.
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func (r Router) ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/lnurlp/{user}", withCORS(r.lnurlHandler.Handle))
	mux.HandleFunc("/.well-known/nostr.json", withCORS(r.nostrHandler.Handle))
	mux.HandleFunc("/.well-known/lnurlp/{user}/callback", withCORS(r.lnurlInvoiceHandler.Handle))
	return mux
}

func (r Router) ListenAndServe(addr string) error {
	slog.Info("Listening on", "addr", addr)
	return http.ListenAndServe(addr, r.ServeMux())
}
