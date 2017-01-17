package components

import (
	"github.com/spf13/cobra"
	"fmt"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "WIP, will remove components installed via this tool.",
	Long: `WIP`,
	Run: func (cmd *cobra.Command, args []string) {
		// Need to happen in order if auto-installing
		gCloudUninstallCmd.Run(cmd, args)
		virtualBoxUninstallCmd.Run(cmd, args)
		minikubeUninstallCmd.Run(cmd, args)
		helmUninstallCmd.Run(cmd, args)
		fmt.Println("Argo has successfully installed or verified all components!")
	},
	// TODO - Uninstall function
}

func init() {
	ComponentsCmd.AddCommand(uninstallCmd)

	uninstallCmd.AddCommand(gCloudUninstallCmd)
	uninstallCmd.AddCommand(virtualBoxUninstallCmd)
	uninstallCmd.AddCommand(minikubeUninstallCmd)
	uninstallCmd.AddCommand(helmUninstallCmd)
}
