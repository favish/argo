package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"runtime"
)

// version is set on build via -ldflags (ie go build -ldflags "-X cmd.version=0.1" .)
// strategy from http://stackoverflow.com/questions/11354518/golang-application-auto-build-versioning
var Version string
var Build string

var versionCmd = &cobra.Command{
	Use:   	"version",
	Short: 	"Get the current version of argo.",
	Run: func (cmd *cobra.Command, args []string) {
		if (len(Version) > 0 && len(Build) > 0) {
			fmt.Printf("Version: %s \n", Version);
			fmt.Printf("Build: %s \n", Build);
			fmt.Printf("GOOS: %s, GOARCH=%s", runtime.GOOS, runtime.GOARCH)
		} else {
			color.Yellow("Are you running argo via `go run ...`?  No version detected from build params!")
		}
	},
}