//* `argo project create PATH --webroot=[OPTIONAL WEB ROOT LOCATION] --sync`
// argo project create --webroot=[optional] --sync=if present run sync command as well
//- path default to .
//- path can be repo
//- if is repo, clone
//
//- create a kubernetes context derived from PATH (either cwd, or repo name)
//- set context to active context
//
//- helm install HELM-CHART(from argo.rc)
//- helm will need to be informed which directory to use to mount the project.
//- default $PWD/docroot
//- Optionally Tell helm where the docroot is on this specific machine (developer's box)
//
//- Notify user infrastructure is complete and they need to run argo sync to update database and files
//- or sync after if flag is present

package project

import (
	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"github.com/favish/argo/util"
	"fmt"
	"github.com/spf13/viper"
	"bytes"
	"os"
)

var createCmd = &cobra.Command{
	Use:   	"deploy",
	Short: 	"Deploy argo project.",
	Long: 	`
		Run in directory that contains an argo configuration file (json/yml).
		Creates a kubernetes context and sets up in kubectl.
		Starts and configures the correct chart via Helm.
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if approve := util.GetApproval(fmt.Sprintf("This will create a deployment in the %s environment, are you sure?", projectConfig.GetString("environment"))); !approve {
			color.Yellow("Deployment cancelled by user.")
			return
		}

		if exists := checkExisting(); exists {
			color.Yellow("Project is already running!  Check helm/kubernetes for a running project.  If you want to update this, run `argo project update` instead")
			os.Exit(0)
		}

		if (projectConfig.GetString("environment") == "local") {
			setupLocalEnvironment()
		}

		if err := helmUpgrade(); err != nil {
			color.Red("Error installing chart via helm!")
			return
		}

		color.Cyan(`
.  o ..
o . o o.o
  ...oo      		CONGRATULATIONS!
    __[]__
 __|_o_o_o\__
 \""""""""""/
  \. ..  . /
^^^^^^^^^^^^^^^^^^^^
		`)
		if (projectConfig.GetString("environment") == "local") {
			color.Cyan("Local site available at: http://local.%s.com (when finished spinning up)\n \n", projectConfig.GetString("project_name"))
		}
		color.Green("Your project infrastructure has been created on the %s environment!", projectConfig.GetString("environment"))
		color.Green("This has bootstrapped a kubernetes environment, normal kubectl commands will allow you to interrogate your new infra.")
		color.Yellow("If this is your fist time working with this project, use `argo project sync` to obtain databases and files.")
		color.Yellow("It may take a few moments for the infrastructure to spin up.")
	},
}

func setupLocalEnvironment() {
	setImagePullSecret()
	addEtcHosts()
	createSSLCert()
}

func createSSLCert() {
	// Cert should exist as a secret, so if it's already there continue
	if out, _ := util.ExecCmdChain("kubectl get secret tls-secret 2>&1 >/dev/null | grep 'not found'"); len(out) <= 0 {
		return
	}
	hostname := projectConfig.GetString("environments.local.network.hostname")
	color.Yellow("Generating self-signed HTTPS cert...")
	util.ExecCmdChain(fmt.Sprintf("openssl req -x509 -newkey rsa:2048 -keyout argo-key.pem -out argo-cert.pem -days 365 -nodes -subj '/CN=%s'", hostname))
	util.ExecCmdChain("kubectl create secret tls tls-secret --cert=argo-cert.pem --key=argo-key.pem")
	util.ExecCmd("rm", "argo-cert.pem")
	util.ExecCmd("rm", "argo-key.pem")
}

// Local environments will need to use current developer's gcloud credentials
func setImagePullSecret() {
	// Check if secret already exists by grepping stdout and stderr for not found.  if grep returns ok (with output), skip
	if out, _ := util.ExecCmdChain("kubectl get secret gcr 2>&1 >/dev/null | grep 'not found'"); len(out) <= 0 {
		return
	}

	// Get the email address for the active account
	gcloudConfig := viper.New()
	gcloudConfig.SetConfigType("yaml")
	// Use cmd chain to get stdout back
	output, _ := util.ExecCmdChain("gcloud info --format=yaml")
	outByte := []byte(output)
	gcloudConfig.ReadConfig(bytes.NewBuffer(outByte))
	gcloudEmail := gcloudConfig.GetString("config.account")

	gcloudAccessToken, _, _ := util.ExecCmdOut("gcloud", "auth", "print-access-token")

	if err := util.ExecCmd("kubectl",
		"create",
		"secret",
		"docker-registry",
		"gcr",
		"--docker-server=https://gcr.io",
		"--docker-username=oauth2accesstoken",
		fmt.Sprintf("--docker-password=%s", gcloudAccessToken),
		fmt.Sprintf("--docker-email=%s", gcloudEmail)); err != nil {
		color.Red("%v", err)
	}

}

// Add or update an entry to access this project locally into /etc/hosts
func addEtcHosts() {
	if approve := util.GetApproval("Argo can add an /etc/hosts entry for this project for you, would you like to do this?"); approve {
		color.Cyan("Adding/updating entry to /etc/hosts.  Will require sudo permissions...")
		localAddress := projectConfig.GetString("environments.local.network.hostname")
		util.ExecCmdChain(fmt.Sprintf("sudo sed --in-place '/%s/d' /etc/hosts", localAddress))
		util.ExecCmdChain(fmt.Sprintf("echo \"$(minikube ip) %s\" | sudo tee -a /etc/hosts", localAddress))
	} else {
		color.Cyan("Skipping auto-addition of /etc/hosts entry.  You can map this yourself using `echo \"$(minikube ip) PROJECT NAME\" | sudo tee -a /etc/hosts` if you'd like to.")
	}
}