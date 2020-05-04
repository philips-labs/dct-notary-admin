package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/hashicorp/vault/api"
)

type VaultPasswordGenerator struct {
	client  *api.Client
	options VaultPasswordOptions
}

type VaultPasswordOptions struct {
	Len            *uint `json:"length,omitempty"`
	Digits         *uint `json:"digits,omitempty"`
	Symbols        *uint `json:"symbols,omitempty"`
	AllowUppercase *bool `json:"allow_uppercase,omitempty"`
	AllowRepeat    *bool `json:"allow_repeat,omitempty"`
}

type VaultKeyPassword struct {
	Password string `json:"password,omitempty"`
}

type VaultSecret struct {
	Data interface{} `json:"data, omitempty"`
}

func (g *VaultPasswordGenerator) Generate() (string, error) {
	bytes, err := json.Marshal(g.options)
	if err != nil {
		return "", err
	}

	response, err := g.client.Logical().WriteBytes("gen/password", bytes)
	if err != nil {
		return "", err
	}
	if password, ok := response.Data["value"].(string); ok {
		return password, nil
	}
	return "", errors.New("failed to read password")
}

// NewVaultPasswordGenerator Creates the secrets engine using Hashicorp Vault
func NewVaultPasswordGenerator(client *api.Client, options VaultPasswordOptions) PasswordGenerator {
	return &VaultPasswordGenerator{
		client:  client,
		options: options,
	}
}

type VaultCredentialsManager struct {
	client        *api.Client
	passGenerator PasswordGenerator
}

func NewVaultCredentialsManager(client *api.Client, passGenerator PasswordGenerator) *VaultCredentialsManager {
	return &VaultCredentialsManager{
		client:        client,
		passGenerator: passGenerator,
	}
}

func (v *VaultCredentialsManager) StorePassword(key, password string) error {
	path := path.Join("dctna", "data", "dev", key)
	passwd := VaultKeyPassword{Password: password}
	data, err := json.Marshal(VaultSecret{Data: passwd})
	if err != nil {
		return err
	}
	_, err = v.client.Logical().WriteBytes(path, data)
	return err
}

func (v *VaultCredentialsManager) ReadPassword(key string) (string, error) {
	path := path.Join("dctna", "data", "dev", key)
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return "", err
	}
	if secretData, ok := secret.Data["data"].(map[string]interface{}); ok {
		if passwd, ok := secretData["password"].(string); ok {
			return passwd, nil
		}

	}

	return "", fmt.Errorf("failed to read secret, data in unexpected format")
}
