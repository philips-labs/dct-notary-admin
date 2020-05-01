package secrets_test

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"

	"github.com/philips-labs/dct-notary-admin/lib/secrets"
)

var (
	client *api.Client
)

func init() {
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
	cmd := exec.Command("../../vault/prepare.sh", "dev")
	out, err := cmd.Output()
	if err != nil {
		panic(fmt.Errorf("failed to boot vault instance: %w", err))
	}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "VAULT_") {
			kv := strings.Split(line, "=")
			os.Setenv(kv[0], kv[1])
			fmt.Println("Registering environment variable", kv)
		}
	}
}

func userpassLogin(client *api.Client, username, password string) (string, error) {
	options := map[string]interface{}{
		"password": password,
	}
	path := fmt.Sprintf("auth/userpass/login/%s", username)

	// PUT call to get a token
	secret, err := client.Logical().Write(path, options)
	if err != nil {
		return "", err
	}

	token := secret.Auth.ClientToken
	return token, nil
}

func authenticatedClient() (*api.Client, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	token, err := userpassLogin(client, "dctna", "topsecret")
	if err != nil {
		return client, err
	}
	client.SetToken(token)
	return client, nil
}

func uintPtr(value uint) *uint {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func TestVaultPasswordGenerator(t *testing.T) {
	assert := assert.New(t)

	client, err := authenticatedClient()
	if !assert.NoError(err, "failed to authenticate client") {
		return
	}

	testCases := []struct {
		name    string
		options secrets.VaultPasswordOptions
		expLen  int
		exp     *regexp.Regexp
	}{
		{
			name: "all explicit", expLen: 12, exp: regexp.MustCompile("^[a-z]+$"),
			options: secrets.VaultPasswordOptions{
				Len: uintPtr(12), Digits: uintPtr(0), Symbols: uintPtr(0), AllowUppercase: boolPtr(false), AllowRepeat: boolPtr(true),
			},
		},
		{name: "defaults", expLen: 64, exp: nil},
		{
			name: "lowercase alpha numeric only", expLen: 64, exp: regexp.MustCompile("^[a-z]+$"),
			options: secrets.VaultPasswordOptions{
				AllowUppercase: boolPtr(false), Digits: uintPtr(0), Symbols: uintPtr(0),
			},
		},
		{
			name: "alpha numeric only", expLen: 64, exp: regexp.MustCompile("^[A-Za-z]+$"),
			options: secrets.VaultPasswordOptions{
				Digits: uintPtr(0), Symbols: uintPtr(0),
			},
		},
		{
			name: "alpha numeric and digits only", expLen: 64, exp: regexp.MustCompile("^[A-Za-z\\d]+$"),
			options: secrets.VaultPasswordOptions{
				Digits: uintPtr(5), Symbols: uintPtr(0),
			},
		},
		{
			name: "alpha numeric only shorter Length", expLen: 32, exp: regexp.MustCompile("^[a-zA-Z]+$"),
			options: secrets.VaultPasswordOptions{
				Len: uintPtr(32), Digits: uintPtr(0), Symbols: uintPtr(0),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			pgen := secrets.NewVaultPasswordGenerator(client, tt.options)
			password, err := pgen.Generate()
			if !assert.NoError(err) {
				return
			}
			assert.Len(password, tt.expLen)
			if tt.exp != nil {
				assert.Regexp(tt.exp, password)
			}
		})
	}
}
