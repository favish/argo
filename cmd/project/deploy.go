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
	"path/filepath"
	"errors"
	"github.com/spf13/viper"
	"bytes"
	"path"
)

// Setup flag values that will be bound in init()
var environment string
var remote string

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

		// TODO - start minikube if it's not running and environment is local - MEA

		if approve := util.GetApproval(fmt.Sprintf("This will create a deployment in the %s environment, are you sure?", environment)); !approve {
			return
		}

		name, err := locateProject(args)
		if err != nil {
			color.Red("%v", err)
			return
		}

		setupKubectl(name, environment)

		if exists := checkExisting(name); exists {
			color.Yellow("Project is already running!  Check helm for a running project.")
			return
		}

		if (environment == "local") {
			setImagePullSecret()
		}

		if err := helmInstall(name); err != nil {
			color.Red("Error installing chart via helm!")
			return
		}

		color.Cyan(congratulationsText)
	},
}

// Bind flags
func init () {
	createCmd.Flags().StringVarP(&environment, "environment", "e", "local", "Define which environment to apply argo deployment to. Ex: \"local\", \"dev\", or \"prod\".")
}

// Attempt to create or find project
func locateProject(args []string) (string, error) {
	var projectName string
	var err error
	// If args are provided, use them
	// Otherwise try to find an argo config in the current directory
	if (len(args) > 0) {
		validRepo := strings.Split(args[0], "/")
		if (len(validRepo) != 2) {
			return projectName, errors.New("Invalid git repo provided as argument, make sure it follows pattern 'organization/reponame'!")
		} else {
			projectName = validRepo[1]
			// Return error without attempting to clone if directory exists
			if exists := util.DirectoryExists(projectName); exists {
				return projectName, errors.New("Folder matching git project name exists.  Please re-run this command from that directory.")
			}
			err = cloneProject(projectName, args[0])
		}
	} else {
		// Check for presence of argorc
		noConfig := viper.GetBool("noConfig")
		if noConfig {
			return projectName, errors.New("No argo configuration file found!  Please re-run this command in a project with an argo configuration file in it's root, or specifiy a git repo to clone.")
		} else {
			cwdPath, _ := os.Getwd()
			projectName = filepath.Base(cwdPath)
			color.Cyan("Creating project %s from argo config file...", projectName)
		}
	}
	return projectName, err
}

// Update context/project/etc to match environment
func setupKubectl(name string, environment string) {
	// These are both the same in both remote and local
	contextCluster := "minikube"
	contextUser := "minikube"
	if environment != "local" {
		// TODO - validate argo.yml environments - MEA
		project := viper.GetString(fmt.Sprintf("environments.%s.project", environment))
		computeZone := viper.GetString(fmt.Sprintf("environments.%s.compute-zone", environment))
		cluster := viper.GetString(fmt.Sprintf("environments.%s.cluster", environment))
		color.Cyan("Updating gcloud to use %s-%s cluster credentials...", name, environment)
		util.ExecCmd("gcloud", "config", "set", "project", project)
		util.ExecCmd("gcloud", "config", "set", "compute/zone", computeZone)
		util.ExecCmd("gcloud", "container", "clusters", "get-credentials", cluster)

		contextCluster = fmt.Sprintf("gke_%s_%s_%s", project, computeZone, cluster)
		contextUser = fmt.Sprintf("gke_%s_%s_%s", project, computeZone, cluster)
	}

	// If the namespace does not exist, create one.
	if err := util.ExecCmd("kubectl", "get", "namespace", name); err != nil {
		util.ExecCmd("kubectl", "create", "namespace", name)
		color.Cyan("Created new %s kubernetes namespace.", name)
	}

	// Setup a kubectl context and switch to it
	util.ExecCmd("kubectl", "config", "delete-context", name)
	util.ExecCmd("kubectl", "config", "set-context", name, fmt.Sprintf("--cluster=%s", contextCluster), fmt.Sprintf("--user=%s", contextUser), fmt.Sprintf("--namespace=%s", name))
	util.ExecCmd("kubectl", "config", "use-context", name)
	color.Cyan("Created new %s kubectl context and set to active.", name)

}

// Run a check to see if the project already exists in helm
func checkExisting(name string) bool {
	color.Cyan("Ensuring existing helm project does not exist...")
	projectExists := false
	if out, _ := util.ExecCmdChain(fmt.Sprintf("helm status %s | grep 'STATUS: DEPLOYED'", name)); len(out) > 0 {
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

	if environment == "local" {
		helmValues = append(helmValues, fmt.Sprintf("persistence.webroot=%s", path.Join(viper.GetString("PWD"), viper.GetString("persistence.webroot"))))
		helmValues = append(helmValues, fmt.Sprintf("persistence.database=%s", path.Join(viper.GetString("PWD"), viper.GetString("persistence.database"))))
	} else {
		project := viper.GetString(fmt.Sprintf("environments.%s.project", environment))
		computeZone := viper.GetString(fmt.Sprintf("environments.%s.compute-zone", environment))
		instance := viper.GetString(fmt.Sprintf("environments.%s.mysql.instance", environment))

		database := viper.GetString(fmt.Sprintf("environments.%s.mysql.db", environment))

		mysqlInstance := fmt.Sprintf("%s:%s:%s", project, computeZone, instance)

		helmValues = append(helmValues, fmt.Sprintf("mysql.instance=%s", mysqlInstance))
		helmValues = append(helmValues, fmt.Sprintf("mysql.db=%s", database))

		// Do not push or delete if done - MEA
		// TODO - Set this in argo.yml
		helmValues = append(helmValues, "persistence.webroot=us.gcr.io/favish-general/drupal-8-webroot:10")
	}

	err := util.ExecCmd("helm", "install", "--replace", viper.GetString("chart"), "--name", projectName, "--set", strings.Join(helmValues, ","))
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

var congratulationsText = `
     .  o ..
     o . o o.o
	  ...oo      		CONGRATULATIONS! Your Helm Chart has launched.
	    __[]__   		The list of services is available with "minikube service list"
	 __|_o_o_o\__		You may also want to run "argo project sync" to add your database and files.
	 \""""""""""/
	  \. ..  . /
     ^^^^^^^^^^^^^^^^^^^^
`
