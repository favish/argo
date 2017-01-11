/*
	Base hook for component functionality.
	To include in a parent command, add this command to it, all other commands in this package will be
	added to this as sub-commands.

	This is the only Export from the components package
 */

package components

import (
	"github.com/spf13/cobra"
)

var ComponentsCmd = &cobra.Command{
	Use:   "components",
	Short: "Install or uninstall components.",
	Long: `
		Argo will install or uninstall components, generally via brew.
	`,
}