package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// VersionInfo holds version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    time.Time
}

func (v VersionInfo) String() string {
	sb := new(strings.Builder)
	w := tabwriter.NewWriter(sb, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "\tversion:\t%s\n", v.Version)
	fmt.Fprintf(w, "\tcommit:\t%s\n", v.Commit)
	fmt.Fprintf(w, "\tdate:\t%s\n", v.Date.Format(time.RFC3339))
	w.Flush()
	return sb.String()
}

var (
	version    VersionInfo
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "shows version information",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(sprintVersion(rootCmd))
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// ParseDate parse string using time.RFC3339 format or default to time.Now()
func ParseDate(value string) time.Time {
	if value == "unknown" {
		return time.Now()
	}

	parsedDate, err := time.Parse(time.RFC3339, value)
	if err != nil {
		parsedDate = time.Now()
	}
	return parsedDate
}

func sprintVersion(cmd *cobra.Command) string {
	return fmt.Sprintf("%s\n\n%s", cmd.Short, version)
}
