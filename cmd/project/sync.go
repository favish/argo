package project

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update local copies of persistent data to match remote dev.",
	Long: `
		Sync local copies of remote persistent data (user files/database) to local argo environment.
		TODO
			-- files-only
			-- db-only
	`,
}