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

		name, err := locateProject(args)
		if err != nil {
			color.Red("%v", err)
			return
		}

		setupKubectl(name, environment, true)

		color.Green("Environment set to this project.  You can now use kubectl normally to access this infrastructure.")
	},
}
