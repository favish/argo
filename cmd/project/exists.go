package project

import (
	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"os"
)

var existsCmd = &cobra.Command{
	Use:   	"exists",
	Short: 	"Return error if argo project does not yet exist.",
	Long: 	`
		Created primarily for use in CI.
		Run in directory that contains an argo configuration file (json/yml).
	`,
	Run: func(cmd *cobra.Command, args []string) {
		setKubectlConfig(projectConfig.GetString("environment"))
		if exists := checkExisting(); !exists {
			color.Red("Project not found!")
			os.Exit(1)
		} else {
			color.Green("Sucess!  Project chart has been deployed via helm.")
			os.Exit(0)
		}
	},
}
