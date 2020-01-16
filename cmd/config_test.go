package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

var (
	expSettings = `  remote_server.root_ca:          
  remote_server.skiptlsverify:    true
  remote_server.tls_client_cert:  
  remote_server.tls_client_key:   
  remote_server.url:              https://localhost:4443
  server.listen_addr:             :8086
  server.listen_addr_tls:         :8443
  trust_dir:                      %s
`
	expCfg = "\nconfig:\n" + expSettings
)

func init() {
	viper.AddConfigPath("../.notary")
	initConfig()
}

func getExpectedOutput() string {
	return fmt.Sprintf(expCfg, viper.GetString("trust_dir"))
}

func TestSprintConfig(t *testing.T) {
	assert := assert.New(t)
	cfg := sprintConfig()
	exp := getExpectedOutput()
	assert.Equal(exp, cfg, "Expected configuration to match expectation")
}

func TestConfigCommand(t *testing.T) {
	assert := assert.New(t)
	output, err := executeCommand(rootCmd, "config")
	assert.NoError(err)
	exp := getExpectedOutput()
	assert.Equal(exp, output, "Expected configuration to be outputted in a different format")
}

func TestConfigWithParamsCommand(t *testing.T) {
	assert := assert.New(t)
	output, err := executeCommand(rootCmd, "--listen-addr", ":8888", "--listen-addr-tls", ":9443", "config")
	assert.NoError(err)
	expCfg = strings.Replace(expCfg, ":8086", ":8888", 1)
	expCfg = strings.Replace(expCfg, ":8443", ":9443", 1)
	exp := getExpectedOutput()
	assert.Equal(exp, output, "Expected configuration to be outputted in a different format")
}

func TestResolvePaths(t *testing.T) {
	home := os.Getenv("HOME")
	wd, err := os.Getwd()
	assert.NoError(t, err)

	testCases := []struct {
		key, input, exp string
	}{
		{key: "test_absolute", input: "/var/lib/stuff", exp: "/var/lib/stuff"},
		{key: "test_relative", input: "./lib/stuff", exp: path.Join(wd, "lib", "stuff")},
		{key: "test_home", input: "~/stuff", exp: path.Join(home, "stuff")},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(tt *testing.T) {
			viper.Set(tc.key, tc.input)
			resolveConfigPaths(tc.key)
			assert.Equal(tt, tc.exp, viper.Get(tc.key))
		})
	}
}
