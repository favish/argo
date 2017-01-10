package cmd

import (
	"github.com/spf13/cobra"
)

var installVirtualBox = &cobra.Command{
	Use:   "virtualbox",
	Short: "Installs virtualbox. ",
	Long: `FILL THIS OUT -mf`,
	Run: func(cmd *cobra.Command, args []string) {
		install_brew_cask("VBoxManage", "virtualbox")
	},
}

func init() {
	installCmd.AddCommand(installVirtualBox)
}
