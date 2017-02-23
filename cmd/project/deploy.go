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
	"os"
	"github.com/favish/argo/util"
	"fmt"
	"strings"
	"errors"
	"github.com/spf13/viper"
	"bytes"
	"path"
)

// TODO - Pull project values into struct here or in config for easier re-use - MEA

var createCmd = &cobra.Command{
	Use:   	"deploy",
	Short: 	"Deploy argo project.",
	Long: 	`
		Run in directory that contains an argo configuration file (json/yml).
		Creates a kubernetes context and sets up in kubectl.
		Starts and configures the correct chart via Helm.
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if approve := util.GetApproval(fmt.Sprintf("This will create a deployment in the %s environment, are you sure?", environment)); !approve {
			color.Yellow("Deployment cancelled by user.")
			return
		}

		name, err := locateProject(args)
		if err != nil {
			color.Red("%v", err)
			return
		}

		setupKubectl(name, environment, false)

		if exists := checkExisting(name); exists {
			color.Yellow("Project is already running!  Check helm/kubernetes for a running project.")
			return
		}

		if (environment == "local") {
			setImagePullSecret()
		}

		if err := helmInstall(name); err != nil {
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
		color.Green("Your project infrastructure has been created on the %s environment!", environment)
		color.Green("This has bootstrapped a kubernetes environment, normal kubectl commands will allow you to interrogate your new infra.")
		if (environment == "local") {
			color.Cyan("For minikube, running `minikube service list` will show your minikube IP and nodeports.  Your services are available through there.")
		}
		color.Yellow("If this is your fist time working with this project, use `argo project sync` to obtain databases and files.")
	},
}

// Run a check to see if the project already exists in helm
func checkExisting(name string) bool {
	color.Cyan("Ensuring existing helm project does not exist...")
	projectExists := false
	if out, _ := util.ExecCmdChain(fmt.Sprintf("helm status %s | grep 'STATUS: DEPLOYED'", name)); len(out) > 0 {
		color.Yellow(out)
		projectExists = true
	}
	return projectExists
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

// Accept projectname and chart to install project infra via helm
func helmInstall(projectName string) error {

	color.Cyan("Installing project chart via helm...")

	var helmValues []string

	helmValues = append(helmValues, fmt.Sprintf("namespace=%s", projectName))
	helmValues = append(helmValues, fmt.Sprintf("environment_type=%s", environment))

	// TODO - Blackfire credential management? Currently deploying to both environments - MEA
	helmValues = append(helmValues, fmt.Sprintf("blackfire.server_id=%s", viper.GetString("BLACKFIRE_SERVER_ID")))
	helmValues = append(helmValues, fmt.Sprintf("blackfire.server_token=%s", viper.GetString("BLACKFIRE_SERVER_TOKEN")))

	helmValues = append(helmValues, fmt.Sprintf("php_image=%s", viper.GetString("php-image")))
	helmValues = append(helmValues, fmt.Sprintf("nginx_image=%s", viper.GetString("nginx-image")))
	helmValues = append(helmValues, fmt.Sprintf("web_image=%s", viper.GetString("web-image")))

	if environment == "local" {
		helmValues = append(helmValues, fmt.Sprintf("local.webroot=%s", path.Join(viper.GetString("PWD"), viper.GetString("environments.local.webroot"))))
		helmValues = append(helmValues, fmt.Sprintf("local.project_root=%s", viper.GetString("PWD")))
		helmValues = append(helmValues, fmt.Sprintf("mysql.db=%s", viper.GetString("environments.local.mysql.db")))
		helmValues = append(helmValues, fmt.Sprintf("mysql.pass=%s", viper.GetString("environments.local.mysql.pass")))
		helmValues = append(helmValues, fmt.Sprintf("mysql.user=%s", viper.GetString("environments.local.mysql.user")))
	} else {
		project := viper.GetString(fmt.Sprintf("environments.%s.project", environment))
		computeZone := viper.GetString(fmt.Sprintf("environments.%s.compute-zone", environment))
		instance := viper.GetString(fmt.Sprintf("environments.%s.mysql.instance", environment))

		database := viper.GetString(fmt.Sprintf("environments.%s.mysql.db", environment))

		mysqlInstance := fmt.Sprintf("%s:%s:%s", project, computeZone, instance)

		helmValues = append(helmValues, fmt.Sprintf("mysql.instance=%s", mysqlInstance))
		helmValues = append(helmValues, fmt.Sprintf("mysql.db=%s", database))

		appImage := viper.GetString(fmt.Sprintf("environments.%s.application-image", environment))
		helmValues = append(helmValues, fmt.Sprintf("application.image=%s", appImage))
	}

	command := fmt.Sprintf("helm install --replace %s --name %s --set %s", viper.GetString("chart"), projectName, strings.Join(helmValues, ","))
	out, err := util.ExecCmdChainCombinedOut(command)
	if (err != nil) {
		color.Red(out)
	}
	return err
}

func cloneProject(projectName string, gitRepo string) error {
	gitUrlTpl := "git@github.com:%s.git"
	gitUrl := fmt.Sprintf(gitUrlTpl, gitRepo)

	color.Cyan("Creating project %s from %s", projectName, gitUrl)

	err := util.ExecCmd("git", "clone", gitUrl, projectName)

	os.Chdir(projectName)

	// Reload config from this directory
	if noConfig := viper.ReadInConfig(); noConfig != nil {
		err = errors.New("Cloned a project without an argo configuration file in it's root.  Please add one and run this command again.")
	}

	return err
}
