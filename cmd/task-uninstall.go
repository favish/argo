package cmd

import (
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalls minikube, kubernetes, helm and docker.",
	Long: `Favish Cloud uninstalls and configures your local environemnt with
all of the things you need to develop locally and deploy remotely.`,
}

func init() {
	RootCmd.AddCommand(uninstallCmd)
}
