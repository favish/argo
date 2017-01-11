package components

import (
	"github.com/spf13/cobra"
	"fmt"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install various sub-components needed for Favish local development.",
	Run: func (cmd *cobra.Command, args []string) {
		// Need to happen in order if auto-installing
		gCloudInstallCmd.Run(cmd, args)
		virtualBoxInstallCmd.Run(cmd, args)
		minikubeInstallCmd.Run(cmd, args)
		helmInstallCmd.Run(cmd, args)
		fmt.Println("Argo has successfully installed or verified all components!")
	},
}

func init() {
	installCmd.AddCommand(gCloudInstallCmd)
	installCmd.AddCommand(virtualBoxInstallCmd)
	installCmd.AddCommand(minikubeInstallCmd)
	installCmd.AddCommand(helmInstallCmd)

	ComponentsCmd.AddCommand(installCmd)
}
