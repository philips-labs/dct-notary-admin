package notary

import (
	"context"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestListTargets(t *testing.T) {
	assert := assert.New(t)

	expected := []Key{
		Key{ID: "b635efeddff59751e8b6b59abb45383555103d702e7d3f46fbaaa9a8ac144ab8", GUN: "docker.io/marcofranssen/whalesay", Role: "targets"},
		Key{ID: "d22b2a4c0651b833f0b1a536068c5ba8588041abe7d058aab95fffc5b78c98bd", GUN: "docker.io/marcofranssen/nginx", Role: "targets"},
		Key{ID: "b45192be4389bac3f49f8feeee2aefc478b36cab1c9f56574d7e29e452fc0185", GUN: "docker.io/marcofranssen/openjdk", Role: "targets"},
	}

	expandedConfigPath, err := homedir.Expand("~/.notary/config.json")
	assert.NoError(err)

	s, err := NewService(expandedConfigPath)
	assert.NoError(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	targets, err := s.ListTargets(ctx)
	assert.NoError(err)
	assert.Len(targets, len(expected))
	assert.ElementsMatch(expected, targets)
}
