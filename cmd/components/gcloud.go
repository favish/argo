package components

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
)

var gCloudInstallCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Installs google cloud sdk (gcloud) and kubektl. ",
	Run: func (cmd *cobra.Command, args []string) {
		// Install gcloud and kubectl
		util.BrewCaskInstall("gcloud", "google-cloud-sdk")
	},
}

var gCloudUninstallCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Uninstalls google cloud sdk (gcloud) and kubektl. ",
	Run: func (cmd *cobra.Command, args []string) {
		util.BrewCaskUninstall("gcloud", "google-cloud-sdk")
	},
}

func gCloudAutoCompleteRemove () {
	// Remove line with 'google-cloud-sdk' and add completions where necessary
	util.ExecCmd("sed","-i.pre-gcloud.bak", "/Caskroom\\/google-cloud-sdk/,1 d", util.Home + "/.zshrc")
}

func gCloudAutocompleteRemove() {
	util.AppendToFile(util.Home + "/.zshrc", "\n# Google Cloud SDK autocompletions in Caskroom/google-cloud-sdk\n" +
		"source '/usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/path.zsh.inc'\n" +
		"source '/usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/completion.zsh.inc'")
}
