//* `argo project create PATH --webroot=[OPTIONAL WEB ROOT LOCATION] --sync`
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
)

var createCmd = &cobra.Command{
	Use:   	"create",
	Short: 	"Create/initialize argo project.",
	Long: 	`
		Use either a github repo or a directory containing a github repo (default '.'/$CWD).
		Creates new kubectl context and sets it to the current active context.
		Install the chart via helm.
	`,
}
