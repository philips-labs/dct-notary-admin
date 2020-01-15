package lib

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"github.com/philips-labs/dct-notary-admin/lib/notary"
)

const (
	//ContentTypePlainText holds HTTP Content-Type text/plain
	ContentTypePlainText = "text/plain; charset=utf-8"
)

// Server provides a http.Server with graceful shutdown.
type Server struct {
	l           *zap.Logger
	redirectTLS *http.Server
	*http.Server
}

// NewServer creates a Server serving application endpoints
//
// The server implements a graceful shutdown and utilizes zap.Logger to log Requests.
func NewServer(c *ServerConfig, n *notary.Service, l *zap.Logger) *Server {
	l.Info("Configuring server")
	r := configureAPI(n, l)

	errorLog, _ := zap.NewStdLogAt(l, zap.ErrorLevel)
	srvRedirectTLS := http.Server{
		Addr: c.ListenAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, _ := net.SplitHostPort(r.Host)
			u := r.URL
			u.Host = net.JoinHostPort(host, c.ListenAddrTLS[1:])
			u.Scheme = "https"
			http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		}),
		ErrorLog:     errorLog,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	srv := http.Server{
		Addr:         c.ListenAddrTLS,
		Handler:      r,
		ErrorLog:     errorLog,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{l, &srvRedirectTLS, &srv}
}

// Start runs ListenAndServe on the http.Server with graceful shutdown
func (srv *Server) Start() {
	srv.l.Info("Starting server...")
	defer srv.l.Sync()

	go func() {
		if err := srv.redirectTLS.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srv.l.Fatal("Could not listen on", zap.String("addr", srv.redirectTLS.Addr), zap.Error(err))
		}
	}()
	go func() {
		if err := srv.ListenAndServeTLS("certs/server.crt", "certs/server.key"); err != nil && err != http.ErrServerClosed {
			srv.l.Fatal("Could not listen on", zap.String("addr", srv.Addr), zap.Error(err))
		}
	}()
	srv.l.Info("Server is ready to handle requests", zap.String("addr", srv.redirectTLS.Addr), zap.String("addrTLS", srv.Addr))
	srv.gracefullShutdown()
}

func (srv *Server) gracefullShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	srv.l.Info("Server is shutting down", zap.String("reason", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv.redirectTLS.SetKeepAlivesEnabled(false)
	srv.SetKeepAlivesEnabled(false)
	if err := srv.redirectTLS.Shutdown(ctx); err != nil {
		srv.l.Fatal("Could not gracefully shutdown the server", zap.Error(err))
	}
	if err := srv.Shutdown(ctx); err != nil {
		srv.l.Fatal("Could not gracefully shutdown the server", zap.Error(err))
	}
	srv.l.Info("Server stopped")
}
