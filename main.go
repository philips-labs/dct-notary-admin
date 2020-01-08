package main

import (
	"flag"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mitchellh/go-homedir"

	"github.com/philips-labs/dct-notary-admin/notary"
)

var (
	listenAddr       string
	listenAddrTLS    string
	notaryConfigFile string
)

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":8086", "server listen address")
	flag.StringVar(&listenAddrTLS, "listen-addr-tls", ":8443", "server tls listen address")
	flag.StringVar(&notaryConfigFile, "notary-config-file", "~/.notary/config.json", "path to the configuration file to use")
	flag.Parse()

	logger, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel))
	if err != nil {
		log.Fatalf("Can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	expandedNotaryConfigFile, err := homedir.Expand(notaryConfigFile)
	if err != nil {
		logger.Fatal("failed to expand home directory", zap.Error(err))
	}

	config := &Config{
		ListenAddr:       listenAddr,
		ListenAddrTLS:    listenAddrTLS,
		NotaryConfigFile: expandedNotaryConfigFile,
	}

	n, err := notary.NewService(config.NotaryConfigFile)
	if err != nil {
		logger.Fatal("failed to create notary service", zap.Error(err))
	}

	server := NewServer(config, n, logger)
	server.Start()
}
