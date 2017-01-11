package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/cmd/util"
)

var helmInstallCmd = &cobra.Command{
	Use:   "helm",
	Short: "Installs helm. ",
	Run:  func (cmd *cobra.Command, args []string) {
		util.InstallBrew("helm", "helm")
	},
}