package notary

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedTargets = []Key{
		Key{ID: "b635efeddff59751e8b6b59abb45383555103d702e7d3f46fbaaa9a8ac144ab8", GUN: "docker.io/marcofranssen/whalesay", Role: "targets"},
		Key{ID: "d22b2a4c0651b833f0b1a536068c5ba8588041abe7d058aab95fffc5b78c98bd", GUN: "docker.io/marcofranssen/nginx", Role: "targets"},
		Key{ID: "b45192be4389bac3f49f8feeee2aefc478b36cab1c9f56574d7e29e452fc0185", GUN: "docker.io/marcofranssen/openjdk", Role: "targets"},
	}
	expectedSigner = Key{ID: "eb9dd99255f91efeba139941fbfdb629f11c2353704de07a2ad653d22311c88b", Role: "marcofranssen"}
)

func TestListTargets(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewService(&NotaryConfig{
		TrustDir: "~/.docker/trust",
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	})
	targets, err := s.ListTargets(ctx)
	assert.NoError(err)
	assert.Len(targets, len(expectedTargets))
	assert.ElementsMatch(expectedTargets, targets)
}

func TestListDelegates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewService(&NotaryConfig{
		TrustDir: "~/.docker/trust",
		RemoteServer: RemoteServerConfig{
			URL:           "https://localhost:4443",
			SkipTLSVerify: true,
		},
	})

	for _, target := range expectedTargets {
		t.Run(target.GUN, func(tt *testing.T) {
			assert := assert.New(tt)
			delegates, err := s.ListDelegates(ctx, &target)
			assert.NoError(err)
			assert.Len(delegates, 1)
			assert.Len(delegates[expectedSigner.Role], 1)
			assert.Equal(expectedSigner, delegates[expectedSigner.Role][0])
		})
	}
}
