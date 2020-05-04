package notary

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/passphrase"
	"github.com/theupdateframework/notary/tuf/data"
)

func GetPassphraseRetriever() notary.PassRetriever {
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

// CleanupKeys utility function for tests to cleanup test generated data
func CleanupKeys(trustStore string, keys ...string) error {
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

// KeyFilter allows to filter keys using this filter
type KeyFilter func(Key) bool

var (
	// TargetsFilter filters all keys by role equals targets
	TargetsFilter KeyFilter = RoleFilter("targets")
	// SnapshotsFilter filters all keys by role equals targets
	SnapshotsFilter KeyFilter = RoleFilter("snapshot")
	// RootFilter filters all root keys by role equals root
	RootFilter KeyFilter = RoleFilter("root")
)

// IDFilter filters the keys by the given id
func IDFilter(id string) KeyFilter {
	return KeyFilter(func(k Key) bool { return strings.HasPrefix(k.ID, id) })
}

// RoleFilter filters the keys by Role
func RoleFilter(role string) KeyFilter {
	return func(k Key) bool {
		return k.Role == role
	}
}

// GUNFilter filters the keys by GUN
func GUNFilter(gun string) KeyFilter {
	return func(k Key) bool {
		return k.GUN == strings.Trim(gun, " \t")
	}
}

// AndFilter combines all filters using a AND operation, where all filters must evaluate true
func AndFilter(filters ...KeyFilter) KeyFilter {
	return func(k Key) bool {
		for _, f := range filters {
			if f(k) == false {
				return false
			}
		}
		return true
	}
}

// KeyChanToSlice transforms a channel to a Slice
func KeyChanToSlice(keysChan <-chan Key) []Key {
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

// DelegationPath prefixes 'targets/' to a given roleName
func DelegationPath(roleName string) data.RoleName {
	return data.RoleName(path.Join(data.CanonicalTargetsRole.String(), roleName))
}
