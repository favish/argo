package components

import (
	"github.com/spf13/cobra"
	"github.com/fatih/color"
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
		kubectlInstallCmd.Run(cmd, args)
		color.Green("Argo has successfully installed or verified all components.")
		color.Yellow("You may want to run `argo components start` to bootstrap your local environment!")
	},
}

func init() {
	installCmd.AddCommand(gCloudInstallCmd)
	installCmd.AddCommand(virtualBoxInstallCmd)
	installCmd.AddCommand(minikubeInstallCmd)
	installCmd.AddCommand(helmInstallCmd)
	installCmd.AddCommand(kubectlInstallCmd)

	ComponentsCmd.AddCommand(installCmd)
}
