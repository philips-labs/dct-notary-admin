package notary

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustmanager"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
)

const (
	releasedRoleName = "Repo Admin"
)

var (
	releasesRole = data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "releases"))
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
		URL           string `json:"url"`
		RootCA        string `json:"root_ca"`
		TLSClientKey  string `json:"tls_client_key"`
		TLSClientCert string `json:"tls_client_cert"`
		SkipTLSVerify bool   `json:"skipTLSVerify"`
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

func (s *Service) ListDelegates(ctx context.Context, target *Key) (map[string][]Key, error) {
	var delegates map[string][]Key
	if target == nil {
		return delegates, nil
	}

	gun := data.GUN(target.GUN)
	rt, err := getTransport(s.config, gun, readOnly)
	if err != nil {
		return nil, err
	}

	repo, err := client.NewFileCachedRepository(
		s.config.TrustDir,
		gun,
		s.config.RemoteServer.URL,
		rt,
		getPassphraseRetriever(),
		trustpinning.TrustPinConfig{})

	if err != nil {
		return nil, err
	}
	delegationRoles, err := repo.GetDelegationRoles()
	if err != nil {
		return delegates, err
	}
	delegates = getDelegationRoleToKeyMap(delegationRoles)

	return delegates, err
}

func getDelegationRoleToKeyMap(rawDelegationRoles []data.Role) map[string][]Key {
	signerRoleToKeyIDs := make(map[string][]Key)
	for _, delRole := range rawDelegationRoles {
		switch delRole.Name {
		case releasesRole, data.CanonicalRootRole, data.CanonicalSnapshotRole, data.CanonicalTargetsRole, data.CanonicalTimestampRole:
			continue
		default:
			keys := make([]Key, len(delRole.KeyIDs))
			signer := notaryRoleToSigner(delRole.Name)
			for i, key := range delRole.KeyIDs {
				keys[i] = Key{ID: key, Role: signer}
			}
			signerRoleToKeyIDs[signer] = keys
		}
	}
	return signerRoleToKeyIDs
}

func notaryRoleToSigner(tufRole data.RoleName) string {
	//  don't show a signer for "targets" or "targets/releases"
	if isReleasedTarget(data.RoleName(tufRole.String())) {
		return releasedRoleName
	}
	return strings.TrimPrefix(tufRole.String(), "targets/")
}

func isReleasedTarget(role data.RoleName) bool {
	return role == data.CanonicalTargetsRole || role == releasesRole
}
