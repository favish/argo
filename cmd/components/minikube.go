package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
)

var minikubeInstallCmd = &cobra.Command{
	Use:   "minikube",
	Short: "Installs minikube.",
	Run: func(cmd *cobra.Command, args []string) {
		// Install gcloud and kubectl
		util.BrewCaskInstall("minikube", "minikube")
		util.ExecCmd("minikube", "addons", "enable", "ingress")
		//TODO - make this source minikube env - mf
		// Remove line with 'minikube docker-env' and add completions where necessary
		util.ExecCmd("sed","-i.pre-minikube.bak", "/minikube docker-env/,1 d", util.Home + "/.zshrc")
		util.AppendToFile(util.Home + "/.zshrc", "\n# Point docker to minikube with 'minikube docker-env'\n" +
			"eval $(minikube docker-env)\n")
	},
}

var minikubeUninstallCmd = &cobra.Command{
	Use:   "minikube",
	Short: "Uninstalls minikube.",
	Run: func(cmd *cobra.Command, args []string) {
		// Install gcloud and kubectl
		util.BrewCaskUninstall("minikube", "minikube")
		//TODO - make this source minikube env - mf

	},
}

func autoCompleteSomething () {
	// Remove line with 'minikube docker-env' and add completions where necessary
	util.ExecCmd("sed","-i.pre-minikube.bak", "/minikube docker-env/,1 d", util.Home + "/.zshrc")
	util.AppendToFile(util.Home + "/.zshrc", "\n# Point docker to minikube with 'minikube docker-env'\n" +
		"eval $(minikube docker-env)\n")
}
