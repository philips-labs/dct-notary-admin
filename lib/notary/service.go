package notary

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/storage"
	"github.com/theupdateframework/notary/trustmanager"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/signed"
)

const (
	releasedRoleName = "Repo Admin"
)

var (
	releasesRole = data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "releases"))
	// ErrGunMandatory when no GUN is given this error is thrown
	ErrGunMandatory = fmt.Errorf("must specify a GUN")
	// ErrPublicKeysAndPathsMandatory when no Public Keys and Paths are provided
	ErrPublicKeysAndPathsMandatory = fmt.Errorf("public key(s) and path(s) are required")
)

// Key holds Path and GUN to keys
type Key struct {
	ID   string `json:"id"`
	GUN  string `json:"gun,omitempty"`
	Role string `json:"role"`
}

// KeyData holds private, public keydata
type KeyData struct {
	Key  []byte        `json:"key,omitempty"`
	Role data.RoleName `json:"role,omitempty"`
	GUN  data.GUN      `json:"gun,omitempty"`
}

// TUFMetadata holds tuf metadata
type TUFMetadata struct {
	Root      *data.SignedRoot
	Targets   map[data.RoleName]*data.SignedTargets
	Snapshot  *data.SignedSnapshot
	Timestamp *data.SignedTimestamp
}

// Service notary service exposes notary operations
type Service struct {
	config         *Config
	storageFactory func() (trustmanager.KeyStore, error)
	retriever      notary.PassRetriever
	log            *zap.Logger
}

// NewService creates a new notary service object
func NewService(config *Config, storageFactory func() (trustmanager.KeyStore, error), passRetriever notary.PassRetriever, log *zap.Logger) *Service {
	return &Service{config, storageFactory, passRetriever, log}
}

// CreateRepository creates a new repository with the given id
func (s *Service) CreateRepository(ctx context.Context, cmd CreateRepoCommand) error {
	if err := cmd.GuardHasGUN(); err != nil {
		return err
	}
	sanitizedGUN := cmd.SanitizedGUN()

	fact := ConfigureRepo(s.config, s.retriever, true, readWrite)
	nRepo, err := fact(sanitizedGUN)
	if err != nil {
		return err
	}

	rootKeyIDs, err := importRootKey(s.log, cmd.RootKey, nRepo, s.retriever)
	if err != nil {
		return err
	}

	rootCerts, err := importRootCert(cmd.RootCert)
	if err != nil {
		return err
	}

	// if key is not defined but cert is, then clear the key to allow key to be searched in keystore
	if cmd.RootKey == "" && cmd.RootCert != "" {
		rootKeyIDs = []string{}
	}

	if err = nRepo.InitializeWithCertificate(rootKeyIDs, rootCerts); err != nil {
		return err
	}

	return maybeAutoPublish(s.log, cmd.AutoPublish, sanitizedGUN, s.config, s.retriever)
}

// DeleteRepository deletes the repository for the given gun
func (s *Service) DeleteRepository(ctx context.Context, cmd DeleteRepositoryCommand) error {
	if err := cmd.GuardHasGUN(); err != nil {
		return err
	}
	sanitizedGUN := cmd.SanitizedGUN()
	// Only initialize a roundtripper if we get the remote flag
	var err error
	var rt http.RoundTripper
	var remoteDeleteInfo string
	if cmd.DeleteRemote {
		rt, err = getTransport(s.config, sanitizedGUN, admin)
		if err != nil {
			return err
		}
		remoteDeleteInfo = " and remote"
	}

	if err := client.DeleteTrustData(
		s.config.TrustDir,
		sanitizedGUN,
		s.config.RemoteServer.URL,
		rt,
		cmd.DeleteRemote,
	); err != nil {
		return err
	}
	s.log.Info(fmt.Sprintf("Successfully deleted local%s trust data for repository", remoteDeleteInfo), zap.Stringer("gun", sanitizedGUN))
	return nil
}

// AddDelegation add a new delegate key to the specified repository target
func (s *Service) AddDelegation(ctx context.Context, cmd AddDelegationCommand) error {
	if err := cmd.GuardHasGUN(); err != nil {
		return err
	}
	if len(cmd.DelegationKeys) == 0 || len(cmd.Paths) == 0 {
		return ErrPublicKeysAndPathsMandatory
	}
	sanitizedGUN := cmd.SanitizedGUN()

	fact := ConfigureRepo(s.config, s.retriever, false, readWrite)
	nRepo, err := fact(sanitizedGUN)
	if err != nil {
		return err
	}

	err = nRepo.AddDelegation(DelegationPath("releases"), cmd.DelegationKeys, cmd.Paths)
	if err != nil {
		return fmt.Errorf("failed to create delegation: %w", err)
	}
	err = nRepo.AddDelegation(cmd.Role, cmd.DelegationKeys, cmd.Paths)
	if err != nil {
		return fmt.Errorf("failed to create delegation: %w", err)
	}
	err = nRepo.AddDelegation(DelegationPath("releases"), cmd.DelegationKeys, cmd.Paths)
	if err != nil {
		return fmt.Errorf("failed to create delegation: %w", err)
	}

	return maybeAutoPublish(s.log, cmd.AutoPublish, sanitizedGUN, s.config, s.retriever)
}

// RemoveDelegation remove a delegation from specified GUN
func (s *Service) RemoveDelegation(ctx context.Context, cmd RemoveDelegationCommand) error {
	if err := cmd.GuardHasGUN(); err != nil {
		return err
	}
	sanitizedGUN := cmd.SanitizedGUN()
	fact := ConfigureRepo(s.config, s.retriever, false, readWrite)
	nRepo, err := fact(sanitizedGUN)
	if err != nil {
		return err
	}

	err = nRepo.RemoveDelegationKeys(DelegationPath("releases"), []string{cmd.KeyID})
	if err != nil {
		return fmt.Errorf("failed to create delegation: %w", err)
	}
	err = nRepo.RemoveDelegationKeys(cmd.Role, []string{cmd.KeyID})
	if err != nil {
		return fmt.Errorf("failed to create delegation: %w", err)
	}
	return maybeAutoPublish(s.log, cmd.AutoPublish, sanitizedGUN, s.config, s.retriever)
}

// StreamKeys returns a Stream of Key
func (s *Service) StreamKeys(ctx context.Context) (<-chan Key, error) {
	keysChan := make(chan Key, 2)
	storage, err := s.storageFactory()
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(keysChan)
		keys := storage.ListKeys()
		for keyID, keyInfo := range keys {
			keysChan <- Key{ID: keyID, Role: keyInfo.Role.String(), GUN: keyInfo.Gun.String()}
		}
	}()

	return keysChan, nil
}

// ListRootKeys lists all the notary root keys
func (s *Service) ListRootKeys(ctx context.Context) ([]Key, error) {
	return s.ListKeys(ctx, RootFilter)
}

// ListTargets lists all the notary target keys
func (s *Service) ListTargets(ctx context.Context) ([]Key, error) {
	return s.ListKeys(ctx, TargetsFilter)
}

// ListKeys lists all the notary keys filtered by the given filter
func (s *Service) ListKeys(ctx context.Context, filter KeyFilter) ([]Key, error) {
	keysChan, err := s.StreamKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}
	filteredChan := Reduce(ctx, keysChan, filter)
	filtered := KeyChanToSlice(filteredChan)

	return filtered, nil
}

// GetTargetByGUN retrieves a target by its GUN
func (s *Service) GetTargetByGUN(ctx context.Context, gun data.GUN) (*Key, error) {
	targetKeys, err := s.ListKeys(ctx, AndFilter(TargetsFilter, GUNFilter(gun.String())))
	if err != nil && len(targetKeys) != 1 {
		return nil, err
	}
	return &targetKeys[0], nil
}

// GetKeyByID retrieves a by its id
func (s *Service) GetKeyByID(ctx context.Context, id string) (*Key, error) {
	if len(id) < 7 {
		return nil, fmt.Errorf("you must provide at least 7 characters of the path: %w", ErrInvalidID)
	}

	keysChan, err := s.StreamKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}
	filteredChan := Reduce(ctx, keysChan, IDFilter(id))

	key, open := <-filteredChan
	if !open {
		return nil, nil
	}
	return &key, nil
}

// ListDelegates returns delegate keys for the given target
func (s *Service) ListDelegates(ctx context.Context, target *Key) (map[string][]Key, error) {
	var delegates map[string][]Key
	delegationRoles, err := s.getTargetDelegationRoles(ctx, target)
	if err != nil {
		return nil, err
	}
	delegates = getDelegationRoleToKeyMap(delegationRoles)

	return delegates, err
}

// GetDelegation retrieves a single delegation
func (s *Service) GetDelegation(ctx context.Context, target *Key, role data.RoleName, keyID string) (*Key, error) {
	delegationRoles, err := s.getTargetDelegationRoles(ctx, target)
	if err != nil {
		return nil, err
	}

	for _, delRole := range delegationRoles {
		switch delRole.Name {
		case releasesRole, data.CanonicalRootRole, data.CanonicalSnapshotRole, data.CanonicalTargetsRole, data.CanonicalTimestampRole:
			continue
		default:
			if delRole.Name == role {
				signer := notaryRoleToSigner(delRole.Name)
				for _, delKeyID := range delRole.KeyIDs {
					key := &Key{ID: delKeyID, Role: signer}
					if IDFilter(keyID)(*key) {
						return key, nil
					}
				}
			}
		}
	}
	return nil, nil
}

// FetchMetadata fetches the TUF metadata for a given GUN.
func (s *Service) FetchMetadata(ctx context.Context, gun data.GUN) (*TUFMetadata, error) {
	rt, err := getTransport(s.config, gun, readOnly)
	if err != nil {
		return nil, err
	}

	trustPinning := trustpinning.TrustPinConfig{}

	repo, err := client.NewFileCachedRepository(
		s.config.TrustDir,
		gun,
		s.config.RemoteServer.URL,
		rt,
		s.retriever,
		trustPinning)

	remote, err := storage.NewNotaryServerStore(s.config.RemoteServer.URL, gun, rt)
	if err != nil {
		return nil, err
	}

	tufRepo, _, err := client.LoadTUFRepo(client.TUFLoadOptions{
		GUN:                    repo.GetGUN(),
		TrustPinning:           trustPinning,
		CryptoService:          repo.GetCryptoService(),
		RemoteStore:            remote,
		AlwaysCheckInitialized: false,
	})
	if err != nil {
		return nil, err
	}

	return &TUFMetadata{
		tufRepo.Root,
		tufRepo.Targets,
		tufRepo.Snapshot,
		tufRepo.Timestamp,
	}, nil
}

// FetchKeys fetches the keys for a given GUN.
func (s *Service) FetchKeys(ctx context.Context, gun data.GUN) (map[string]KeyData, error) {
	rt, err := getTransport(s.config, gun, readOnly)
	if err != nil {
		return nil, err
	}

	storage, err := s.storageFactory()
	if err != nil {
		return nil, err
	}

	repo, err := client.NewFileCachedRepository(
		s.config.TrustDir,
		gun,
		s.config.RemoteServer.URL,
		rt,
		s.retriever,
		trustpinning.TrustPinConfig{})

	if err != nil {
		return nil, err
	}
	cs := repo.GetCryptoService()

	rootKeyIDs := cs.ListKeys(data.CanonicalRootRole)
	targetKeyIDs := cs.ListKeys(data.CanonicalTargetsRole)
	snapshotKeyIDs := cs.ListKeys(data.CanonicalSnapshotRole)

	keys := make(map[string]KeyData)
	for _, v := range rootKeyIDs {
		key, err := s.fetchKeys(cs, v, "")
		if err != nil {
			return nil, err
		}
		keys[v] = *key
	}
	for _, v := range targetKeyIDs {
		ki, err := storage.GetKeyInfo(v)
		if err != nil {
			return nil, err
		}
		if ki.Gun == gun {
			key, err := s.fetchKeys(cs, v, gun)
			if err != nil {
				return nil, err
			}
			keys[v] = *key
		}
	}
	for _, v := range snapshotKeyIDs {
		ki, err := storage.GetKeyInfo(v)
		if err != nil {
			return nil, err
		}
		if ki.Gun == gun {
			key, err := s.fetchKeys(cs, v, gun)
			if err != nil {
				return nil, err
			}
			keys[v] = *key
		}
	}
	return keys, nil
}

func (s *Service) fetchKeys(cs signed.CryptoService, keyID string, gun data.GUN) (*KeyData, error) {
	_, role, err := cs.GetPrivateKey(keyID)
	if err != nil {
		return nil, err
	}
	privKeyFile, err := os.Open(filepath.Join(s.config.TrustDir, "private", keyID+".key"))
	if err != nil {
		return nil, err
	}
	defer privKeyFile.Close()
	privKeyFileInfo, _ := privKeyFile.Stat()
	data := make([]byte, privKeyFileInfo.Size())

	buffer := bufio.NewReader(privKeyFile)
	_, err = buffer.Read(data)

	return &KeyData{
		Key:  data,
		Role: role,
		GUN:  gun,
	}, err
}

func (s *Service) getTargetDelegationRoles(ctx context.Context, target *Key) ([]data.Role, error) {
	if target == nil {
		return nil, nil
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
		s.retriever,
		trustpinning.TrustPinConfig{})

	if err != nil {
		return nil, err
	}
	return repo.GetDelegationRoles()
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
	// don't show a signer for "targets" or "targets/releases"
	if isReleasedTarget(data.RoleName(tufRole.String())) {
		return releasedRoleName
	}
	return strings.TrimPrefix(tufRole.String(), "targets/")
}

func isReleasedTarget(role data.RoleName) bool {
	return role == data.CanonicalTargetsRole || role == releasesRole
}
