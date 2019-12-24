package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
)

const (
	//ContentTypePlainText holds HTTP Content-Type text/plain
	ContentTypePlainText = "text/plain; charset=utf-8"
)

// Server provides a http.Server with graceful shutdown.
type Server struct {
	l *zap.Logger
	*http.Server
}

// NewServer creates a Server serving application endpoints
//
// The server implements a graceful shutdown and utilizes zap.Logger to log Requests.
func NewServer(listenAddr string, l *zap.Logger) *Server {
	l.Info("Configuring server")
	r := configureAPI(l)

	errorLog, _ := zap.NewStdLogAt(l, zap.ErrorLevel)
	srv := http.Server{
		Addr:         listenAddr,
		Handler:      r,
		ErrorLog:     errorLog,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{l, &srv}
}

// Start runs ListenAndServe on the http.Server with graceful shutdown
func (srv *Server) Start() {
	srv.l.Info("Starting server...")
	defer srv.l.Sync()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srv.l.Fatal("Could not listen on", zap.String("addr", srv.Addr), zap.Error(err))
		}
	}()
	srv.l.Info("Server is ready to handle requests", zap.String("addr", srv.Addr))
	srv.gracefullShutdown()
}

func (srv *Server) gracefullShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	srv.l.Info("Server is shutting down", zap.String("reason", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv.SetKeepAlivesEnabled(false)
	if err := srv.Shutdown(ctx); err != nil {
		srv.l.Fatal("Could not gracefully shutdown the server", zap.Error(err))
	}
	srv.l.Info("Server stopped")
}
