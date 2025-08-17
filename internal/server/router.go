package server

import (
	"lmt/internal/app"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Router struct {
	lnurlInvoiceHandler app.LNURLInvoiceHandler
	lnurlHandler        app.LNURLHandler
	nostrHandler        app.NostrHandler
	server              *http.Server
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
	mux.HandleFunc("/api/stop", r.Stop)
	return mux
}

func (r Router) ListenAndServe(addr string) error {
	slog.Info("Listening on", "addr", addr)
	r.server = &http.Server{
		Addr:    addr,
		Handler: r.ServeMux(),
	}
	return r.server.ListenAndServe()
}

func (r Router) Stop(w http.ResponseWriter, req *http.Request) {
	slog.Info("Received stop request, shutting down server...")
	w.WriteHeader(http.StatusOK)
	go func() {
		time.Sleep(2 * time.Second)
		slog.Info("Shutting down server...")
		os.Exit(0)
	}()
}
