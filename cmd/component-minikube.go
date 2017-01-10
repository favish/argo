package cmd

import (
	"github.com/spf13/cobra"
)

var installMinikube = &cobra.Command{
	Use:   "minikube",
	Short: "Installs minikube.",
	Long: `FILL THIS OUT -mf`,
	Run: func(cmd *cobra.Command, args []string) {

		// Install gcloud and kubectl
		install_brew_cask("minikube", "minikube")
                //TODO make this source minikube env

		// Remove line with 'minikube docker-env' and add completions where necessary
		exec_command("sed","-i.pre-minikube.bak", "/minikube docker-env/,1 d", home + "/.zshrc")
		append(home + "/.zshrc", "\n# Point docker to minikube with 'minikube docker-env'\n" +
			"eval $(minikube docker-env)\n")
	},
}

func init() {
	installCmd.AddCommand(installMinikube)
}
