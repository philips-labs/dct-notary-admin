package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

var (
	expSettings = `  remote_server.url:       https://localhost:4443
  server.listen_addr:      :8086
  server.listen_addr_tls:  :8443
  trust_dir:               .
`
	expCfg = "\nconfig:\n" + expSettings
)

func init() {
	viper.AddConfigPath("../.notary")
	initConfig()
}

func TestConfigPrinting(t *testing.T) {
	sb := new(strings.Builder)
	writeSettings(sb, viper.AllSettings())

	assert.Equal(t, expSettings, sb.String(), "Expected printed configuration to match expectation")
}

func TestSprintConfig(t *testing.T) {
	cfg := sprintConfig()

	assert.Equal(t, expCfg, cfg, "Expected configuration to match expectation")
}

func TestConfigCommand(t *testing.T) {
	assert := assert.New(t)
	output, err := executeCommand(rootCmd, "config")
	assert.NoError(err)
	assert.Equal(expCfg, output, "Expected configuration to be outputted in a different format")
}

func TestConfigWithParamsCommand(t *testing.T) {
	assert := assert.New(t)
	output, err := executeCommand(rootCmd, "--listen-addr", ":8888", "--listen-addr-tls", ":9443", "config")
	assert.NoError(err)
	expCfg = strings.Replace(expCfg, ":8086", ":8888", 1)
	expCfg = strings.Replace(expCfg, ":8443", ":9443", 1)
	assert.Equal(expCfg, output, "Expected configuration to be outputted in a different format")
}
