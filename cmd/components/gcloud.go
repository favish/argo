package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/cmd/util"
)

var gCloudInstallCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Installs google cloud sdk (gcloud) and kubektl. ",
	Run: func (cmd *cobra.Command, args []string) {
		// Install gcloud and kubectl
		util.InstallBrew("gcloud", "google-cloud-sdk")
		util.ExecCmd("gcloud", "--no-user-output-enabled", "components", "install", "kubectl")

		// Remove line with 'google-cloud-sdk' and add completions where necessary
		util.ExecCmd("sed","-i.pre-gcloud.bak", "/Caskroom\\/google-cloud-sdk/,1 d", util.Home + "/.zshrc")
	},
}

var gCloudUninstallCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Uninstalls google cloud sdk (gcloud) and kubektl. ",
	//Run: func (cmd *cobra.Command, args []string) {},
}

// TODO - Finish this - MEA
//func gCloudAddAutocomplete() {
//	append(home + "/.zshrc", "\n# Google Cloud SDK autocompletions in Caskroom/google-cloud-sdk\n" +
//		"source '/usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/path.zsh.inc'\n" +
//		"source '/usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/completion.zsh.inc'")
