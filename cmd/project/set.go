package project

import (
	"github.com/spf13/cobra"
	"github.com/fatih/color"
)

var setEnvCmd = &cobra.Command{
	Use:   	"set-env",
	Short: 	"Set your local environment to use this project.",
	Long: 	`
		Run in directory that contains an argo configuration file (json/yml).
		Creates a kubernetes context and sets up in kubectl.
	`,
	Run: func(cmd *cobra.Command, args []string) {

		// Every project command sets up the local environment
		// TODO - Make this create kubectl/gcloud configuration contexts and activate them for usage outside of argo

		color.Green("Environment set to this project.  You can now use kubectl normally to access this infrastructure.")
	},
}
