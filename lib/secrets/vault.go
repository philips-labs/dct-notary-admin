package secrets

import (
	"encoding/json"
	"errors"

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
