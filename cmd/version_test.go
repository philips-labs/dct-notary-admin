package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDateUnknown(t *testing.T) {
	d := ParseDate("unknown")

	assert.WithinDuration(t, time.Now(), d, 1*time.Millisecond)
}

func TestParseDateRFC3339(t *testing.T) {
	exp := "2019-11-17T16:11:22Z"
	d := ParseDate(exp)

	assert.Equal(t, exp, d.Format(time.RFC3339))
}

func TestParseDateNonRFC3339(t *testing.T) {
	d := ParseDate("01 Jan 15 10:00 UTC")
	exp := time.Now()

	assert.Equal(t, exp.Format(time.RFC3339), d.Format(time.RFC3339))
}

func TestVersionCommands(t *testing.T) {
	version = VersionInfo{
		Version: "test",
		Commit:  "ab23f6",
		Date:    ParseDate("2019-11-17T16:11:22Z"),
	}
	exp := fmt.Sprintf(`%s

  version:  %s
  commit:   %s
  date:     %s

`, rootCmd.Short, version.Version, version.Commit, version.Date.Format(time.RFC3339))
	args := []string{"version", "--version", "-v"}

	for _, arg := range args {
		t.Run(arg, func(tt *testing.T) {
			output, err := executeCommand(rootCmd, arg)
			assert := assert.New(tt)
			assert.NoError(err)
			assert.Equal(exp, output)
		})
	}
}
