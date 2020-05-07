package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
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
  vault.addr:                     http://localhost:8200
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

func TestConfigPathRelativeToCwd(t *testing.T) {
	home, err := homedir.Dir()
	assert.NoError(t, err)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	testCases := []struct {
		key, input, exp string
	}{
		{key: "rel_wd_test_empty", input: "", exp: ""},
		{key: "rel_wd_test_absolute", input: "/var/lib/stuff", exp: "/var/lib/stuff"},
		{key: "rel_wd_test_relative", input: "./lib/stuff", exp: filepath.Join(wd, "lib", "stuff")},
		{key: "rel_wd_test_home", input: "~/stuff", exp: filepath.Join(home, "stuff")},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(tt *testing.T) {
			viper.Set(tc.key, tc.input)
			resolveConfigPathsRelativeToCwd(tc.key)
			assert.Equal(tt, tc.exp, viper.Get(tc.key))
		})
	}
}

func TestConfigPathRelativeToConfig(t *testing.T) {
	home, err := homedir.Dir()
	assert.NoError(t, err)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	testCases := []struct {
		key, input, exp string
	}{
		{key: "rel_config_test_empty", input: "", exp: ""},
		{key: "rel_config_test_absolute", input: "/var/lib/stuff", exp: "/var/lib/stuff"},
		{key: "rel_config_test_relative", input: "./lib/stuff", exp: filepath.Join(wd, "../.notary", "lib", "stuff")},
		{key: "rel_config_test_home", input: "~/stuff", exp: filepath.Join(home, "stuff")},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(tt *testing.T) {
			viper.Set(tc.key, tc.input)
			resolveConfigPathsRelativeToConfig(tc.key)
			assert.Equal(tt, tc.exp, viper.Get(tc.key))
		})
	}
}

func TestUnmarshalServerConfig(t *testing.T) {
	assert := assert.New(t)

	cfg, err := unmarshalServerConfig()
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal(":8086", cfg.ListenAddr)
	assert.Equal(":8443", cfg.ListenAddrTLS)
}

func TestUnmarshalNotaryConfig(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	assert.NoError(err)

	cfg, err := unmarshalNotaryConfig()
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal(filepath.Join(wd, "../.notary"), cfg.TrustDir)
	assert.Equal("", cfg.RemoteServer.RootCA)
	assert.True(cfg.RemoteServer.SkipTLSVerify)
	assert.Equal("", cfg.RemoteServer.TLSClientCert)
	assert.Equal("", cfg.RemoteServer.TLSClientKey)
	assert.Equal("https://localhost:4443", cfg.RemoteServer.URL)
}

func TestUnmarshalVaultConfig(t *testing.T) {
	assert := assert.New(t)

	cfg, err := unmarshalVaultConfig()
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.Equal("http://localhost:8200", cfg.Address)
}
