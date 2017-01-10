package cmd

import (
	"github.com/spf13/cobra"
)

var gCloudInstallCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Installs google cloud sdk (gcloud) and kubektl. ",
	Long: `FILL THIS OUT -mf`,
	Run: gCloudInstall,
}

var gCloudUninstallCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Installs google cloud sdk (gcloud) and kubektl. ",
	Long: `FILL THIS OUT -mf`,
	Run: gCloudUninstall,
}


func gCloudInstall (cmd *cobra.Command, args []string) {
	// Install gcloud and kubectl
	install_brew("gcloud", "google-cloud-sdk")
	exec_command("gcloud", "-q", "components", "install", "kubectl")

	// Remove line with 'google-cloud-sdk' and add completions where necessary
	exec_command("sed","-i.pre-gcloud.bak", "/Caskroom\\/google-cloud-sdk/,1 d", home + "/.zshrc")
}

func gCloudUninstall (cmd *cobra.Command, args []string) {
	// Install gcloud and kubectl
	install_brew("gcloud", "google-cloud-sdk")
	exec_command("gcloud", "-q", "components", "install", "kubectl")

	// Remove line with 'google-cloud-sdk' and add completions where necessary
	exec_command("sed","-i.pre-gcloud.bak", "/Caskroom\\/google-cloud-sdk/,1 d", home + "/.zshrc")
}

func gCloudAddAutocomplete() {
	append(home + "/.zshrc", "\n# Google Cloud SDK autocompletions in Caskroom/google-cloud-sdk\n" +
		"source '/usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/path.zsh.inc'\n" +
		"source '/usr/local/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/completion.zsh.inc'")
}

func init() {
	installCmd.AddCommand(gCloudInstallCmd)
	uninstallCmd.AddCommand(gCloudUninstallCmd)
}

