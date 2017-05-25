package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
	"os"
	"fmt"
	"strconv"
)

var rollbackCmd = &cobra.Command{
	Use: 	"rollback",
	Short: 	"Undo the last applied update.",
	Long: 	"If, after running an argo project update you decide to go backwards, your can run this.  Will also be executed if --rollback-on-failure supplied to argo update",
	Run: func (cmd *cobra.Command, args []string) {
		projectName := projectConfig.GetString("project_name")
		command := fmt.Sprintf("helm list | grep %s | awk {'print $2'} | tr -d '\n'", projectName)
		currentRevisionStr, _ := util.ExecCmdChain(command)
		currentRevision, _ := strconv.Atoi(currentRevisionStr)
		rollBackRevision := currentRevision - 1
		rollBackRevisionStr := strconv.Itoa(rollBackRevision)

		if err := util.ExecCmd("helm", "rollback", projectName, rollBackRevisionStr); err != nil {
			color.Red("Rollback failed to apply.", err)
			os.Exit(1)
		}
	},
}