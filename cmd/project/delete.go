package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
	"fmt"
)

var deleteCmd = &cobra.Command{
	Use:   	"delete",
	Short: 	"Delete argo project created via `argo deploy`.",
	Run: func (cmd *cobra.Command, args []string) {
		name := projectConfig.GetString("project-name")
		environment := projectConfig.GetString("environment")

		if approve := util.GetApproval(fmt.Sprintf("This will delete your project's infrastucture for your %s environment are you sure?", environment)); approve {
			util.ExecCmd("helm", "delete", name)
			util.ExecCmd("kubectl", "config", "delete-context", name)
			color.Green("Project infrastructure removed!")
		} else {
			return
		}
	},
}