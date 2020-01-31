package notary

import (
	"strings"

	"github.com/theupdateframework/notary/tuf/data"
)

// TargetCommand holds the target parameters
type TargetCommand struct {
	GUN data.GUN
}

// CreateRepoCommand holds data to create a new repository for the given data.GUN
type CreateRepoCommand struct {
	TargetCommand
	RootKey     string
	RootCert    string
	AutoPublish bool
}

// DeleteRepositoryCommand holds data to delete the repository for the given data.GUN
type DeleteRepositoryCommand struct {
	TargetCommand
	DeleteRemote bool
}

// GuardHasGUN guards that a valid GUN has been provided
func (cmd TargetCommand) GuardHasGUN() error {
	if strings.Trim(cmd.GUN.String(), " \t") == "" {
		return ErrGunMandatory
	}
	return nil
}
