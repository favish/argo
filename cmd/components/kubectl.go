package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
)

var kubectlInstallCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Install kubectl.",
	Run:  func (cmd *cobra.Command, args []string) {
		util.BrewInstall("kubectl", "kubectl")
	},
}

var kubectlUninstallCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Uninstall kubectl.",
	Run:  func (cmd *cobra.Command, args []string) {
		util.BrewUninstall("kubectl", "kubectl")
	},
}
