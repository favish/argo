package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
)

var virtualBoxInstallCmd = &cobra.Command{
	Use:   "virtualbox",
	Short: "Installs virtualbox. ",
	Run: func (cmd *cobra.Command, args []string) {
		util.BrewCaskInstall("VBoxManage", "virtualbox")
	},
}

var virtualBoxUninstallCmd = &cobra.Command{
	Use:   "virtualbox",
	Short: "Installs virtualbox. ",
	Run: func (cmd *cobra.Command, args []string) {
		util.BrewCaskUninstall("VBoxManage", "virtualbox")
	},
}
