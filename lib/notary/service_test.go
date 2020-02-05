package notary

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/cryptoservice"
	"github.com/theupdateframework/notary/trustmanager"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/utils"
)

var (
	trustStore      = "../../.notary"
	rootKeyID       = "760e57b96f72ed27e523633d2ffafe45ae0ff804e78dfc014a50f01f823d161d"
	expectedTargets = []Key{
		Key{ID: "4ea1fec36392486d4bd99795ffc70f3ffa4a76185b39c8c2ab1d9cf5054dbbc9", GUN: "localhost:5000/dct-notary-admin", Role: "targets"},
	}
	expectedSigner = Key{ID: "eb9dd99255f91efeba139941fbfdb629f11c2353704de07a2ad653d22311c88b", Role: "marcofranssen"}
	service        *Service
	fact           RepoFactory
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

	gun := "localhost:5000/test-create-remove/dct-notary-admin"
	createCmd := CreateRepoCommand{TargetCommand: TargetCommand{GUN: data.GUN(gun)}, AutoPublish: true}
	err := service.CreateRepository(ctx, createCmd)
	assert.NoError(err)

	targetKeys, err := service.ListKeys(ctx, AndFilter(TargetsFilter, GUNFilter(gun)))
	assert.NoError(err)

	snapshotKeys, err := service.ListKeys(ctx, AndFilter(SnapshotsFilter, GUNFilter(gun)))
	assert.NoError(err)

	deleteCmd := DeleteRepositoryCommand{TargetCommand: TargetCommand{GUN: data.GUN(gun)}, DeleteRemote: true}
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

	cmd := AddDelegationCommand{TargetCommand: TargetCommand{GUN: data.GUN("localhost:5000/dct-notary-admin")}}
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

	role := data.RoleName("test-add-delegation")
	gun := data.GUN("localhost:5000/dct-notary-admin")
	signerRole := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), role.String()))

	nRepo, err := fact(gun)
	defer nRepo.RemoveDelegationRole(signerRole)

	delgKey := createDelgKey(assert, nRepo, role, true)
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

	for _, target := range expectedTargets {
		t.Run(target.GUN, func(tt *testing.T) {
			assert := assert.New(tt)
			delegates, err := service.ListDelegates(ctx, &target)
			assert.NoError(err)

			assert.Len(delegates, 0)
			// TODO: improve test data
			// if assert.Len(delegates, 1) {
			// 	assert.Len(delegates[expectedSigner.Role], 1)
			// 	assert.Equal(expectedSigner, delegates[expectedSigner.Role][0])
			// }
		})
	}
}

func createDelgKey(assert *assert.Assertions, repo client.Repository, role data.RoleName, x509 bool) data.PublicKey {
	fileKeyStore, err := trustmanager.NewKeyFileStore(trustStore, getPassphraseRetriever())
	if !assert.NoError(err) {
		return nil
	}
	privKey, err := utils.GenerateKey(data.ECDSAKey)
	fileKeyStore.AddKey(trustmanager.KeyInfo{Role: role}, privKey)

	return data.PublicKeyFromPrivate(privKey)
}
