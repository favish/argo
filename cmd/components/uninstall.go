package components

import (
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "WIP, will remove components installed via this tool.",
	Long: `WIP`,
	// TODO - Uninstall function
}

func init() {
	ComponentsCmd.AddCommand(uninstallCmd)
}
