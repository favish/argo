package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start or verify that all components required to operate argo projects are running.",
	Run: func (cmd *cobra.Command, args []string) {
		// Return on commands that exit with errors.  Commands will print stderr for us.

		// Check if minikube is running, and start if it is not.  Oddly `minikube start` does not do this itself - MEA
		if out, _ := util.ExecCmdChain("minikube status | grep 'localkube: Running'"); len(out) <= 0 {
			if err := util.ExecCmd("minikube", "start"); err != nil {
				return
			}
		}

		// Update docker environment to point at minikube
		if yes := util.GetApproval("In order to init helm, we'll need to update your Docker environment variables to use Minikube, OK?"); !yes {
			color.Red("Good day sir!")
			return
		}
		if _, err := util.ExecCmdChain("eval $(minikube docker-env)"); err != nil {
			return
		}

		if err := util.ExecCmd("helm", "init"); err != nil {
			return
		}

		color.Green("You're now ready to start using argo projects!")
	},
}

// Export startCmd publicly so other commands can run it
var StartCmd = startCmd;

func init() {
	ComponentsCmd.AddCommand(startCmd)
}
