package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"os"
	"fmt"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update local copies of persistent data to match remote dev.",
	Long: `
		Sync local copies of remote persistent data (user files/database) to local argo environment.
	`,
}

var filesCmd = &cobra.Command{
	Use: "files",
	Short: "Sync files FROM > TO.",
	Long: `Use this command to retrieve or push your files to and from remote environments.
	Ex. "argo project sync files dev local" brings dev files to your machine.
	Using rsync and project argo.yml configuration under the hood.
	Valid choices for FROM/TO are local, prod, or dev.
	`,
	Run: func(cmd *cobra.Command, args []string) {

		validateArgs(args)

		from, to := getSyncPaths(args)

		if approve := util.GetApproval(fmt.Sprintf("This will sync files from %s (%s) to %s (%s), proceed?", args[0], from, args[1], to)); !approve {
			os.Exit(1)
		} else {
			color.Cyan("In order to update this project's infra, updating your shell context to point to it...")
			setupKubectl(ProjectName, environment, true)

			color.Yellow("Adding ssh aliases via gcloud for compute instances...")

			util.ExecCmd("gcloud", "compute", "config-ssh")

			util.ExecCmd("rsync", "-avz", from, to)
		}
	},
}

func init() {
	syncCmd.AddCommand(filesCmd)
}

func validateArgs(args []string) {
	if (len(args) > 2) {
		color.Red("Too many args.  Command accepts only FROM and TO arguments")
	}

	validArgs := map[string]bool {
		"dev": true,
		"local": true,
		"prod": true,
	}
	// Validate arguments
	for _, arg := range args {
		if !validArgs[arg] {
			color.Red("Invalid argument provided.  Must be one of: local, prod, dev.")
			os.Exit(1)
		}
	}
}

func getSyncPaths(args []string) (string, string) {

	locations := map[string]string {
		"dev": fmt.Sprintf("%s:%s%s/*", viper.GetString("environments.dev.files-instance"), "/mnt/disks/", ProjectName),
		"prod": fmt.Sprintf("%s:%s%s/*", viper.GetString("environments.prod.files-instance"), "/mnt/disks/", ProjectName),
		"local": "./docroot/sites/default/files/",
	}

	from := locations[args[0]]
	to := locations[args[1]]

	return from, to
}