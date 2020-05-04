package secrets

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
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

func uintPtr(value uint) *uint {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func TestVaultPasswordGenerator(t *testing.T) {
	assert := assert.New(t)

	client, err := NewAuthenticatedVaultClient("dctna", "topsecret")
	if !assert.NoError(err, "failed to authenticate client") {
		return
	}

	testCases := []struct {
		name    string
		options VaultPasswordOptions
		expLen  int
		exp     *regexp.Regexp
	}{
		{
			name: "all explicit", expLen: 12, exp: regexp.MustCompile("^[a-z]+$"),
			options: VaultPasswordOptions{
				Len: uintPtr(12), Digits: uintPtr(0), Symbols: uintPtr(0), AllowUppercase: boolPtr(false), AllowRepeat: boolPtr(true),
			},
		},
		{name: "defaults", expLen: 64, exp: nil},
		{
			name: "lowercase alpha numeric only", expLen: 64, exp: regexp.MustCompile("^[a-z]+$"),
			options: VaultPasswordOptions{
				AllowUppercase: boolPtr(false), Digits: uintPtr(0), Symbols: uintPtr(0),
			},
		},
		{
			name: "alpha numeric only", expLen: 64, exp: regexp.MustCompile("^[A-Za-z]+$"),
			options: VaultPasswordOptions{
				Digits: uintPtr(0), Symbols: uintPtr(0),
			},
		},
		{
			name: "alpha numeric and digits only", expLen: 64, exp: regexp.MustCompile("^[A-Za-z\\d]+$"),
			options: VaultPasswordOptions{
				Digits: uintPtr(5), Symbols: uintPtr(0),
			},
		},
		{
			name: "alpha numeric only shorter Length", expLen: 32, exp: regexp.MustCompile("^[a-zA-Z]+$"),
			options: VaultPasswordOptions{
				Len: uintPtr(32), Digits: uintPtr(0), Symbols: uintPtr(0),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			pgen := NewVaultPasswordGenerator(client, tt.options)
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

func TestStoreKeyPassword(t *testing.T) {
	assert := assert.New(t)

	client, err := NewAuthenticatedVaultClient("dctna", "topsecret")
	if !assert.NoError(err, "failed to authenticate vault client") {
		return
	}

	cm := NewVaultCredentialsManager(client, NewVaultPasswordGenerator(client, VaultPasswordOptions{}), zap.NewNop())
	err = cm.StorePassword("localhost:5000/dctna", "super secret", "")
	assert.NoError(err)
}

func TestReadSecret(t *testing.T) {
	assert := assert.New(t)

	client, err := NewAuthenticatedVaultClient("dctna", "topsecret")
	if !assert.NoError(err) {
		return
	}

	cm := NewVaultCredentialsManager(client, NewVaultPasswordGenerator(client, VaultPasswordOptions{}), zap.NewNop())
	err = cm.StorePassword("760e57b96f72ed27e523633d2ffafe45ae0ff804e78dfc014a50f01f823d161d", "test1234", "root")
	if !assert.NoError(err) {
		return
	}

	t.Run("get existing secret", func(t *testing.T) {
		secret, err := cm.ReadPassword("760e57b96f72ed27e523633d2ffafe45ae0ff804e78dfc014a50f01f823d161d")

		assert.NoError(err)
		assert.NotNil(secret)
		assert.Equal("test1234", secret.Password)
		assert.Equal("root", secret.Alias)
	})

	t.Run("get non existing secret", func(t *testing.T) {
		passwd, err := cm.ReadPassword("unknown-secret")

		assert.Error(err)
		assert.IsType(ErrNotFound, errors.Unwrap(err))
		assert.Nil(passwd)
	})
}
