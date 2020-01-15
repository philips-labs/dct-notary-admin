package main

import (
	"github.com/philips-labs/dct-notary-admin/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	v := cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    cmd.ParseDate(date),
	}
	cmd.Execute(v)
}
