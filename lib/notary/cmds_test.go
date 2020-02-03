package notary

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theupdateframework/notary/tuf/data"
)

func TestGuardHasGun(t *testing.T) {
	testCases := []struct {
		gun           string
		shouldSucceed bool
	}{
		{"", false},
		{" ", false},
		{"\t ", false},
		{"\t\t", false},
		{"  \n", false},
		{"\r\n\t", false},
		{"abc", true},
		{"localhost:5000/dct-notary-admin", true},
		{"localhost:5000/dct-notary-admin\t", true},
		{"localhost:5000/dct-notary-admin\n", true},
		{"  localhost:5000/dct-notary-admin", true},
		{"  localhost:5000/dct-notary-admin\r\n", true},
		{"  localhost:5000/   dct-notary-admin\r\n", true},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(tt *testing.T) {
			cmd := TargetCommand{GUN: data.GUN(tc.gun)}
			err := cmd.GuardHasGUN()
			if tc.shouldSucceed {
				assert.NoError(tt, err)
			} else {
				assert.Error(tt, err)
			}
		})
	}
}
