package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
	"fmt"
)

var updateCmd = &cobra.Command{
	Use:   	"update",
	Short: 	"Update running argo project created via `argo deploy`.",
	Run: func (cmd *cobra.Command, args []string) {

		color.Cyan("In order to update this project's infra, updating your shell context to point to it...")

		if approve := util.GetApproval(fmt.Sprintf("This will apply updated configuration to the %s infrastructure, are you sure?", projectConfig.GetString("environment"))); approve {
			helmUpgrade()
			color.Green("Project updated!")
		} else {
			return
		}
	},
}