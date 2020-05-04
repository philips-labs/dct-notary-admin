package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/hashicorp/vault/api"
	"github.com/theupdateframework/notary"

	"go.uber.org/zap"
)

var ErrNotFound = errors.New("secret not found")

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
	Alias    string `json:"alias,omitempty"`
}

type VaultSecret struct {
	Data interface{} `json:"data, omitempty"`
}

func NewAuthenticatedVaultClient(username, password string) (*api.Client, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	token, err := userpassLogin(client, username, password)
	if err != nil {
		return client, err
	}
	client.SetToken(token)
	return client, nil
}

func userpassLogin(client *api.Client, username, password string) (string, error) {
	options := map[string]interface{}{
		"password": password,
	}
	path := path.Join("auth", "userpass", "login", username)

	// PUT call to get a token
	secret, err := client.Logical().Write(path, options)
	if err != nil {
		return "", err
	}

	token := secret.Auth.ClientToken
	return token, nil
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
	log           *zap.Logger
}

func NewVaultCredentialsManager(client *api.Client, passGenerator PasswordGenerator, log *zap.Logger) *VaultCredentialsManager {
	return &VaultCredentialsManager{
		client:        client,
		passGenerator: passGenerator,
		log:           log,
	}
}

func (v *VaultCredentialsManager) PassRetriever() notary.PassRetriever {
	maxRetries := 3
	return func(keyName, alias string, createNew bool, numAttempts int) (string, bool, error) {
		log := v.log.With(
			zap.String("keyName", keyName),
			zap.String("alias", alias),
			zap.Bool("createNew", createNew),
			zap.Int("numAttempts", numAttempts),
		)

		log.Debug("getting credential")
		secret, err := v.ReadOrGenerate(keyName, alias, createNew)
		if err != nil || secret == nil {
			log.Error("failed to get password", zap.Error(err))
			return "", numAttempts > maxRetries, fmt.Errorf("failed to get credential: %w", err)
		}

		return secret.Password, numAttempts > maxRetries, nil
	}
}

func (v *VaultCredentialsManager) ReadOrGenerate(key, alias string, createNew bool) (*VaultKeyPassword, error) {
	var secret *VaultKeyPassword
	secret, err := v.ReadPassword(key)
	if err != nil {
		if errors.Is(err, ErrNotFound) && createNew {
			v.log.Debug("generating new credential")
			passwd, err := v.Generate()
			if err != nil {
				return nil, err
			}
			v.log.Debug("persisting new credential")
			err = v.StorePassword(key, passwd, alias)
			if err != nil {
				return nil, err
			}
			secret = &VaultKeyPassword{Password: passwd, Alias: alias}
		}
		return nil, err
	}
	return secret, nil
}

func (v *VaultCredentialsManager) Generate() (string, error) {
	return v.passGenerator.Generate()
}

func (v *VaultCredentialsManager) StorePassword(key, password, alias string) error {
	path := path.Join("dctna", "data", "dev", key)
	passwd := VaultKeyPassword{Password: password, Alias: alias}
	data, err := json.Marshal(VaultSecret{Data: passwd})
	if err != nil {
		return err
	}
	_, err = v.client.Logical().WriteBytes(path, data)
	return err
}

func (v *VaultCredentialsManager) ReadPassword(key string) (*VaultKeyPassword, error) {
	path := path.Join("dctna", "data", "dev", key)
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, fmt.Errorf("%s: %w", path, ErrNotFound)
	}
	if secretData, ok := secret.Data["data"].(map[string]interface{}); ok {
		if passwd, ok := secretData["password"].(string); ok {
			if alias, ok := secretData["alias"].(string); ok {
				return &VaultKeyPassword{Password: passwd, Alias: alias}, nil
			}
			return &VaultKeyPassword{Password: passwd}, nil
		}
	}

	return nil, fmt.Errorf("failed to read secret, data in unexpected format")
}
