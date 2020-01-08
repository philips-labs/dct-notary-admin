package notary

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/theupdateframework/notary/trustmanager"
)

// Key holds Path and GUN to keys
type Key struct {
	ID   string `json:"id"`
	GUN  string `json:"gun"`
	Role string `json:"role"`
}

// Service notary service exposes notary operations
type Service struct {
	config     *notaryConfig
	configFile string
}

type notaryConfig struct {
	TrustDir     string `json:"trust_dir"`
	RemoteServer struct {
		URL string `json:"url"`
	} `json:"remote_server"`
}

// NewService creates a new notary service object
func NewService(configFile string) (*Service, error) {
	config, err := getConfig(configFile)
	if err != nil {
		return nil, err
	}
	return &Service{config, configFile}, nil
}

func getConfig(configFile string) (*notaryConfig, error) {
	var config notaryConfig
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	expandedTrustDir, err := homedir.Expand(config.TrustDir)
	config.TrustDir = expandedTrustDir
	return &config, nil
}

func (s *Service) StreamKeys(ctx context.Context) (<-chan Key, error) {
	keysChan := make(chan Key, 2)
	fileKeyStore, err := trustmanager.NewKeyFileStore(s.config.TrustDir, getPassphraseRetriever())
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(keysChan)
		keys := fileKeyStore.ListKeys()
		for keyID, keyInfo := range keys {
			keysChan <- Key{ID: keyID, Role: keyInfo.Role.String(), GUN: keyInfo.Gun.String()}
		}
	}()

	return keysChan, nil
}

// ListTargets lists all the notary target keys
func (s *Service) ListTargets(ctx context.Context) ([]Key, error) {
	keysChan, err := s.StreamKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}
	targetChan := Reduce(ctx, keysChan, TargetsFilter)

	targets := make([]Key, 0)
	for target := range targetChan {
		targets = append(targets, target)
	}

	return targets, nil
}

// GetTarget retrieves a target by its path/id
func (s *Service) GetTarget(ctx context.Context, id string) (*Key, error) {
	if len(id) < 7 {
		return nil, fmt.Errorf("you must provide at least 7 characters of the path: %w", ErrInvalidID)
	}

	keysChan, err := s.StreamKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}
	targetChan := Reduce(ctx, keysChan, IDFilter(id))

	target, open := <-targetChan
	if !open {
		return nil, nil
	}
	return &target, nil
}
