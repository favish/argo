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
)

var createCmd = &cobra.Command{
	Use:   	"create",
	Short: 	"Create/initialize argo project.",
	Long: 	`
		Run in directory that contains an argo configuration file (json/yml).
		Creates a kubernetes context and sets up in kubectl.
		Installs and configures the correct chart via Helm.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		name, err := locateProject(args)
		if err != nil {
			color.Red("%v", err)
			return
		}
		if exists := checkExisting(name); exists {
			color.Yellow("Project is already running!  Check helm for a running project.")
			return
		}
		initNamespace(name)
		helmInstall(name, viper.GetString("chart"), viper.GetString("webroot"))

		color.Cyan(congratulationsText)
	},
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

// Run a check to see if the project already exists in helm
func checkExisting(name string) bool {
	projectExists := false
	out, _ := util.ExecCmdChain(fmt.Sprintf("helm status %s | grep 'STATUS: DEPLOYED'", name))
	if len(out) > 0 {
		projectExists = true
	}
	return projectExists
}


// Create the k8s namespace, setup a kubectl context and switch to it
func initNamespace(name string) {
	util.ExecCmd("kubectl", "create", "namespace", name)
	util.ExecCmd("kubectl", "config", "delete-context", name)
	util.ExecCmd("kubectl", "config", "set-context", name, "--cluster=minikube", "--user=minikube", fmt.Sprintf("--namespace=%s", name))
	util.ExecCmd("kubectl", "config", "use-context", name)
	color.Cyan("Created new %s namespace and kubectl context and set it the active kubectl context.", name)
}

// Accept projectname and chart to install project infra via helm
func helmInstall(projectName string, chart string, webroot string) {
	var helmValues []string
	helmValues = append(helmValues, fmt.Sprintf("namespace=%s", projectName))
	helmValues = append(helmValues, fmt.Sprintf("blackfire.server_id=%s", viper.GetString("BLACKFIRE_SERVER_ID")))
	helmValues = append(helmValues, fmt.Sprintf("blackfire.server_token=%s", viper.GetString("BLACKFIRE_SERVER_TOKEN")))
	helmValues = append(helmValues, fmt.Sprintf("webroot=%s", viper.GetString("webroot")))
	util.ExecCmd("helm", "install", "--replace", chart, "--name", projectName, "--set", strings.Join(helmValues, ","))
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
