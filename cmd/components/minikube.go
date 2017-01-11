package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/cmd/util"
)

var minikubeInstallCmd = &cobra.Command{
	Use:   "minikube",
	Short: "Installs minikube.",
	Run: func(cmd *cobra.Command, args []string) {
		// Install gcloud and kubectl
		util.InstallBrewCask("minikube", "minikube")
                //TODO - make this source minikube env - mf
		// Remove line with 'minikube docker-env' and add completions where necessary
		util.ExecCmd("sed","-i.pre-minikube.bak", "/minikube docker-env/,1 d", util.Home + "/.zshrc")
		util.AppendToFile(util.Home + "/.zshrc", "\n# Point docker to minikube with 'minikube docker-env'\n" +
			"eval $(minikube docker-env)\n")
	},
}