package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
)

var helmInstallCmd = &cobra.Command{
	Use:   "helm",
	Short: "Install helm.",
	Run:  func (cmd *cobra.Command, args []string) {
		util.BrewInstall("helm", "kubernetes-helm")
	},
}

var helmUninstallCmd = &cobra.Command{
	Use:   "helm",
	Short: "Uninstall helm.",
	Run:  func (cmd *cobra.Command, args []string) {
		util.BrewUninstall("helm", "kubernetes-helm")
	},
}
