package notary

import (
	"context"
	"os"
	"strings"

	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/passphrase"
	"github.com/theupdateframework/notary/tuf/data"
)

func getPassphraseRetriever() notary.PassRetriever {
	baseRetriever := passphrase.PromptRetriever()
	env := map[string]string{
		"root":       os.Getenv("NOTARY_ROOT_PASSPHRASE"),
		"targets":    os.Getenv("NOTARY_TARGETS_PASSPHRASE"),
		"snapshot":   os.Getenv("NOTARY_SNAPSHOT_PASSPHRASE"),
		"delegation": os.Getenv("NOTARY_DELEGATION_PASSPHRASE"),
	}

	return func(keyName string, alias string, createNew bool, numAttempts int) (string, bool, error) {
		if v := env[alias]; v != "" {
			return v, numAttempts > 1, nil
		}
		// For delegation roles, we can also try the "delegation" alias if it is specified
		// Note that we don't check if the role name is for a delegation to allow for names like "user"
		// since delegation keys can be shared across repositories
		// This cannot be a base role or imported key, though.
		if v := env["delegation"]; !data.IsBaseRole(data.RoleName(alias)) && v != "" {
			return v, numAttempts > 1, nil
		}
		return baseRetriever(keyName, alias, createNew, numAttempts)
	}
}

// KeyFilter allows to filter keys using this filter
type KeyFilter func(Key) bool

var (
	// TargetsFilter filters all keys by role equals targets
	TargetsFilter KeyFilter = func(k Key) bool { return k.Role == "targets" }
	// RootFilter filters all root keys by role equals root
	RootFilter KeyFilter = func(k Key) bool { return k.Role == "root" }
)

// IDFilter filters the keys by the given id
func IDFilter(id string) KeyFilter {
	return KeyFilter(func(k Key) bool { return strings.HasPrefix(k.ID, id) })
}

func KeyChanToList(keysChan <-chan Key) []Key {
	keys := make([]Key, 0)
	for key := range keysChan {
		keys = append(keys, key)
	}
	return keys
}

// Reduce reduces the channel to only stream elements matching the provided filter
func Reduce(ctx context.Context, keysChan <-chan Key, filter KeyFilter) <-chan Key {
	targetChan := make(chan Key)

	go func() {
		defer close(targetChan)
		for {
			select {
			case key, open := <-keysChan:
				if !open {
					return
				}
				if filter(key) {
					targetChan <- key
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return targetChan
}
