//* `argo project create PATH --webroot=[OPTIONAL WEB ROOT LOCATION] --sync`

// argo project create --webroot=[optional] --sync=if present run sync command as well
//- path default to .
//- path can be repo
//- if is repo, clone
//
//- create a kubernetes context derived from PATH (either cwd, or repo name)
//- set context to active context
//
//- helm install HELM-CHART(from argo.rc)
//- helm will need to be informed which directory to use to mount the project.
//- default $PWD/webroot
//
//- Notify user infrastructure is complete and they need to run argo sync to update database and files
//- or sync after if flag is present
package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
)


var createCmd = &cobra.Command{
	Use:   	"create",
	Short: 	"Create/initialize argo project.",
	Long: 	`
		Run in directory that contains an .argorc file in a parent directory.
		Creates a kubernetes context and sets up in kubectl.
		Installs and configures the correct chart via Helm.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for an .argorc in this folder

	},
}
