package main

import (
	"flag"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	listenAddr    string
	listenAddrTLS string
)

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":8086", "server listen address")
	flag.StringVar(&listenAddrTLS, "listen-addr-tls", ":8443", "server tls listen address")
	flag.Parse()

	logger, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel))
	if err != nil {
		log.Fatalf("Can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	server := NewServer(listenAddr, listenAddrTLS, logger)
	server.Start()
}
