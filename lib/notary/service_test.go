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
		Key{ID: "b635efeddff59751e8b6b59abb45383555103d702e7d3f46fbaaa9a8ac144ab8", GUN: "docker.io/marcofranssen/whalesay", Role: "targets"},
		Key{ID: "d22b2a4c0651b833f0b1a536068c5ba8588041abe7d058aab95fffc5b78c98bd", GUN: "docker.io/marcofranssen/nginx", Role: "targets"},
		Key{ID: "b45192be4389bac3f49f8feeee2aefc478b36cab1c9f56574d7e29e452fc0185", GUN: "docker.io/marcofranssen/openjdk", Role: "targets"},
	}
	expectedSigner = Key{ID: "eb9dd99255f91efeba139941fbfdb629f11c2353704de07a2ad653d22311c88b", Role: "marcofranssen"}
)

func TestListRootKeys(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewService(&Config{
		TrustDir: trustStore,
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, zap.NewNop())

	rootKeys, err := s.ListRootKeys(ctx)
	assert.NoError(err)
	assert.Len(rootKeys, 1)
	assert.Equal(rootKeyID, rootKeys[0].ID)
	assert.Equal("", rootKeys[0].GUN)
}

func TestListTargets(t *testing.T) {
	t.Skip("Will need stubbing")
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewService(&Config{
		TrustDir: "./.notary",
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, zap.NewNop())
	targets, err := s.ListTargets(ctx)
	assert.NoError(err)
	assert.Len(targets, len(expectedTargets))
	assert.ElementsMatch(expectedTargets, targets)
}

func TestCreateRepository(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	os.Setenv("NOTARY_ROOT_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_TARGETS_PASSPHRASE", "test1234")
	os.Setenv("NOTARY_SNAPSHOT_PASSPHRASE", "test1234")

	s := NewService(&Config{
		TrustDir: trustStore,
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, zap.NewNop())

	gun := "localhost:5000/dct-notary-admin"
	err := s.CreateRepository(ctx, CreateRepoCommand{GUN: data.GUN(gun)})
	assert.NoError(err)

	targetKeys, err := s.ListTargets(ctx)
	assert.NoError(err)
	assert.Len(targetKeys, 1)
	assert.Equal(gun, targetKeys[0].GUN)

	snapshotKeys, err := s.ListKeys(ctx, SnapshotsFilter)
	assert.NoError(err)
	assert.Len(snapshotKeys, 1)
	assert.Equal(gun, snapshotKeys[0].GUN)

	err = os.Remove(filepath.Join(trustStore, "private", fmt.Sprintf("%s.key", targetKeys[0].ID)))
	assert.NoError(err, "Failed to cleanup newly created target key %s", targetKeys[0].ID)
	err = os.Remove(filepath.Join(trustStore, "private", fmt.Sprintf("%s.key", snapshotKeys[0].ID)))
	assert.NoError(err, "Failed to cleanup newly created snapshot key %s", snapshotKeys[0].ID)
}

func TestListDelegates(t *testing.T) {
	t.Skip("Will need stubbing")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewService(&Config{
		TrustDir: "./.notary",
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	}, zap.NewNop())

	for _, target := range expectedTargets {
		t.Run(target.GUN, func(tt *testing.T) {
			assert := assert.New(tt)
			delegates, err := s.ListDelegates(ctx, &target)
			assert.NoError(err)
			if assert.Len(delegates, 1) {
				assert.Len(delegates[expectedSigner.Role], 1)
				assert.Equal(expectedSigner, delegates[expectedSigner.Role][0])
			}
		})
	}
}
