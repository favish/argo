package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
	"fmt"
	"os"
)

var updateCmd = &cobra.Command{
	Use:   	"update",
	Short: 	"Update running argo project created via `argo deploy`.",
	Run: func (cmd *cobra.Command, args []string) {
		if exists := checkExisting(); !exists {
			color.Red("Project does not exist yet, try running `argo project deploy` instead.")
			os.Exit(1)
		}
		setKubectlConfig(projectConfig.GetString("environment"))

		if approve := util.GetApproval(fmt.Sprintf("This will apply updated configuration to the %s infrastructure, are you sure?", projectConfig.GetString("environment"))); approve {
			helmUpgrade()
			color.Green("Project updated!")
		} else {
			return
		}
	},
}

func init() {
	updateCmd.Flags().Bool("rollback-on-failure", false, "If true, attempt to rollback to recover from helm deployments that result in failures.")
	projectConfig.BindPFlag("rollback-on-failure", updateCmd.Flags().Lookup("rollback-on-failure"))
}