package notary

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/theupdateframework/notary/trustmanager"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/utils"
)

const (
	chars      = "abcdefghijklmnopqrstuvwxyz_"
	trustStore = "../../.notary"
	rootKeyID  = "760e57b96f72ed27e523633d2ffafe45ae0ff804e78dfc014a50f01f823d161d"
)

var (
	service         *Service
	fact            RepoFactory
	expectedTargets = []Key{
		Key{ID: "4ea1fec36392486d4bd99795ffc70f3ffa4a76185b39c8c2ab1d9cf5054dbbc9", GUN: "localhost:5000/dct-notary-admin", Role: "targets"},
	}
)

func init() {
	os.Setenv("NOTARY_ROOT_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_TARGETS_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_SNAPSHOT_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_DELEGATION_PASSPHRASE", "test1234")

	config := &Config{
		TrustDir: trustStore,
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}

	fact = ConfigureRepo(config, getPassphraseRetriever(), true, readOnly)
	service = NewService(config, zap.NewNop())
}

func TestListRootKeys(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootKeys, err := service.ListRootKeys(ctx)
	assert.NoError(err)
	assert.Len(rootKeys, 1)
	assert.Equal(rootKeyID, rootKeys[0].ID)
	assert.Equal("", rootKeys[0].GUN)
	assert.Equal("root", rootKeys[0].Role)
}

func TestListTargets(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	targets, err := service.ListTargets(ctx)
	assert.NoError(err)
	assert.Len(targets, len(expectedTargets))
	assert.ElementsMatch(expectedTargets, targets)
}

func TestGetTarget(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	target, err := service.GetTarget(ctx, "4ea1fec36392486d4bd99795ffc70f3ffa4a76185b39c8c2ab1d9cf5054dbbc9")
	assert.NoError(err)
	assert.NotNil(target)
	assert.Equal(expectedTargets[0].ID, target.ID)
	assert.Equal(expectedTargets[0].GUN, target.GUN)
	assert.Equal(expectedTargets[0].Role, target.Role)
}

func TestCreateRepositoryInvalidGUN(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := CreateRepoCommand{TargetCommand: TargetCommand{GUN: data.GUN("\t ")}, AutoPublish: false}
	err := service.CreateRepository(ctx, cmd)
	assert.EqualError(err, ErrGunMandatory.Error())
}

func TestCreateAndRemoveRepository(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gun := randomGUN()
	createCmd := CreateRepoCommand{TargetCommand: TargetCommand{GUN: gun}, AutoPublish: true}
	err := service.CreateRepository(ctx, createCmd)
	assert.NoError(err)

	targetKeys, err := service.ListKeys(ctx, AndFilter(TargetsFilter, GUNFilter(gun.String())))
	assert.NoError(err)

	snapshotKeys, err := service.ListKeys(ctx, AndFilter(SnapshotsFilter, GUNFilter(gun.String())))
	assert.NoError(err)

	deleteCmd := DeleteRepositoryCommand{TargetCommand: TargetCommand{GUN: gun}, DeleteRemote: true}
	err = service.DeleteRepository(ctx, deleteCmd)
	assert.NoError(err)

	gunKeys := append(targetKeys, snapshotKeys...)

	keyIds := make([]string, len(gunKeys))
	for i, key := range gunKeys {
		keyIds[i] = key.ID
	}
	err = CleanupKeys(trustStore, keyIds...)
	assert.NoError(err)
}

func TestAddDelegateWithoutPublicKeyAndPath(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gun := randomGUN()
	cmd := AddDelegationCommand{TargetCommand: TargetCommand{GUN: gun}}
	err := service.AddDelegation(ctx, cmd)
	assert.EqualError(err, ErrPublicKeysAndPathsMandatory.Error())
}

func TestAddDelegateInvalidGUN(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := AddDelegationCommand{TargetCommand: TargetCommand{GUN: data.GUN("\t \t")}}
	err := service.AddDelegation(ctx, cmd)
	assert.EqualError(err, ErrGunMandatory.Error())
}

func TestAddDelegation(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	role := data.RoleName(randomString(8))
	gun := randomGUN()
	signerRole := DelegationPath(role.String())

	id, err := createTestTarget(ctx, gun)
	if !assert.NoError(err) {
		return
	}
	defer func() {
		err := cleanupTarget(ctx, gun, id)
		assert.NoError(err)
	}()

	nRepo, err := fact(gun)
	defer nRepo.RemoveDelegationRole(signerRole)

	delgKey, err := createDelgKey(role)
	if !assert.NoError(err) {
		return
	}
	defer CleanupKeys(trustStore, delgKey.ID())

	if !assert.NotNil(delgKey) {
		return
	}

	cmd := AddDelegationCommand{
		TargetCommand:  TargetCommand{GUN: data.GUN(gun)},
		Role:           signerRole,
		DelegationKeys: []data.PublicKey{delgKey},
		Paths:          []string{""},
		AutoPublish:    true,
	}
	err = service.AddDelegation(ctx, cmd)
	assert.NoError(err)
}

func TestDeleteRepositoryInvalidGUN(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := DeleteRepositoryCommand{TargetCommand: TargetCommand{GUN: data.GUN(" ")}, DeleteRemote: false}
	err := service.DeleteRepository(ctx, cmd)
	assert.EqualError(err, ErrGunMandatory.Error())
}

func TestListDelegates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	gun := randomGUN()
	id, err := createTestTarget(ctx, gun)
	if !assert.NoError(err) {
		return
	}
	defer func() {
		err := cleanupTarget(ctx, gun, id)
		assert.NoError(err)
	}()

	delID, delName, err := addDelegation(ctx, gun)
	assert.NoError(err)

	target, err := service.GetTarget(ctx, id)
	if !assert.NoError(err) {
		return
	}

	delegates, err := service.ListDelegates(ctx, target)
	assert.NoError(err)

	if assert.Len(delegates, 1) {
		assert.Len(delegates[delName.String()], 1)
		assert.Equal(delID, delegates[delName.String()][0].ID)
		assert.Equal(delName.String(), delegates[delName.String()][0].Role)
	}
}

func createDelgKey(role data.RoleName) (data.PublicKey, error) {
	fileKeyStore, err := trustmanager.NewKeyFileStore(trustStore, getPassphraseRetriever())
	if err != nil {
		return nil, err
	}
	privKey, err := utils.GenerateKey(data.ECDSAKey)
	if err != nil {
		return nil, err
	}
	err = fileKeyStore.AddKey(trustmanager.KeyInfo{Role: role}, privKey)
	if err != nil {
		return nil, err
	}
	return data.PublicKeyFromPrivate(privKey), nil
}

func createTestTarget(ctx context.Context, gun data.GUN) (string, error) {
	err := service.CreateRepository(ctx, CreateRepoCommand{
		TargetCommand: TargetCommand{GUN: gun},
		AutoPublish:   true,
	})
	if err != nil {
		return "", err
	}
	targetKey, err := service.GetTargetByGUN(ctx, gun)
	if err != nil {
		return "", err
	}
	return targetKey.ID, nil
}

func addDelegation(ctx context.Context, gun data.GUN) (string, data.RoleName, error) {
	role := data.RoleName(randomString(8))
	signerRole := DelegationPath(role.String())
	delegationKey, err := createDelgKey(role)
	defer CleanupKeys(trustStore, delegationKey.ID())
	if err != nil {
		return "", data.RoleName(""), nil
	}

	pubKeyID, err := utils.CanonicalKeyID(delegationKey)
	if err != nil {
		return "", data.RoleName(""), fmt.Errorf("can't determine public Key ID: %w", err)
	}

	err = service.AddDelegation(ctx, AddDelegationCommand{
		AutoPublish:    true,
		Role:           signerRole,
		DelegationKeys: []data.PublicKey{delegationKey},
		Paths:          []string{""},
		TargetCommand:  TargetCommand{GUN: gun},
	})
	if err != nil {
		return "", data.RoleName(""), nil
	}
	return pubKeyID, role, nil
}

func cleanupTarget(ctx context.Context, gun data.GUN, keyID string) error {
	snapshotKeys, err := service.ListKeys(ctx, AndFilter(SnapshotsFilter, GUNFilter(gun.String())))
	if err != nil {
		return err
	}
	err = service.DeleteRepository(ctx, DeleteRepositoryCommand{TargetCommand: TargetCommand{GUN: gun}})
	if err != nil {
		return err
	}

	keyIds := make([]string, len(snapshotKeys))
	for i, key := range snapshotKeys {
		keyIds[i] = key.ID
	}
	return CleanupKeys(trustStore, append(keyIds, keyID)...)
}

func randomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[seededRand.Intn(len(chars))]
	}
	return string(b)
}

func randomGUN() data.GUN {
	return data.GUN(fmt.Sprintf("localhost:5000/random-test-guns/dctna-%s", randomString(10)))
}
