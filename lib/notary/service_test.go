package notary

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/theupdateframework/notary/tuf/data"
)

var (
	trustStore      = "../../.notary"
	rootKeyID       = "760e57b96f72ed27e523633d2ffafe45ae0ff804e78dfc014a50f01f823d161d"
	expectedTargets = []Key{
		Key{ID: "c3b49d8c15f339864a21c90a0b7c242e737e6a8a4d1ad73603bfdf0709f01241", GUN: "localhost:5000/dct-notary-admin", Role: "targets"},
	}
	expectedSigner = Key{ID: "eb9dd99255f91efeba139941fbfdb629f11c2353704de07a2ad653d22311c88b", Role: "marcofranssen"}
	service        *Service
)

func init() {
	os.Setenv("NOTARY_ROOT_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_TARGETS_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_SNAPSHOT_PASSPHRASE", "test1234")

	service = NewService(&Config{
		TrustDir: trustStore,
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, zap.NewNop())
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

func TestCreateAndRemoveRepository(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gun := "localhost:5000/test-create-remove/dct-notary-admin"
	err := service.CreateRepository(ctx, CreateRepoCommand{GUN: data.GUN(gun), AutoPublish: true})
	assert.NoError(err)

	targetKeys, err := service.ListKeys(ctx, AndFilter(TargetsFilter, NotGUNFilter(expectedTargets[0].GUN)))
	assert.NoError(err)
	assert.Len(targetKeys, 1)
	assert.Equal(gun, targetKeys[0].GUN)

	snapshotKeys, err := service.ListKeys(ctx, AndFilter(SnapshotsFilter, NotGUNFilter(expectedTargets[0].GUN)))
	assert.NoError(err)
	assert.Len(snapshotKeys, 1)
	assert.Equal(gun, snapshotKeys[0].GUN)

	err = service.DeleteRepository(ctx, DeleteRepositoryCommand{GUN: data.GUN(gun), DeleteRemote: true})
	assert.NoError(err)

	err = cleanupKeys(targetKeys[0].ID, snapshotKeys[0].ID)
	assert.NoError(err)
}

func TestListDelegates(t *testing.T) {
	t.Skip("Will need stubbing")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, target := range expectedTargets {
		t.Run(target.GUN, func(tt *testing.T) {
			assert := assert.New(tt)
			delegates, err := service.ListDelegates(ctx, &target)
			assert.NoError(err)
			if assert.Len(delegates, 1) {
				assert.Len(delegates[expectedSigner.Role], 1)
				assert.Equal(expectedSigner, delegates[expectedSigner.Role][0])
			}
		})
	}
}

func cleanupKeys(keys ...string) error {
	failures := make([]string, 0)
	for _, key := range keys {
		err := os.Remove(filepath.Join(trustStore, "private", fmt.Sprintf("%s.key", key)))
		if err != nil {
			failures = append(failures, key)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("failed to remove keys %s", failures)
	}
	return nil
}
