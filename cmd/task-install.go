package cmd

import (
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs minikube, kubernetes, helm and docker.",
	Long: `Favish Cloud installs and configures your local environemnt with
all of the things you need to develop locally and deploy remotely.`,
	Run: func(cmd *cobra.Command, args []string) {
		gCloudInstall(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
