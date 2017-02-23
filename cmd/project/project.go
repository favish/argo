package project

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/fatih/color"
	"github.com/favish/argo/util"
	"github.com/favish/argo/cmd/components"
	"os"
	"path/filepath"
	"strings"
	"errors"
	"fmt"
	"path"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage individual argo projects.",
	Long: `Manipulate argo projects.`,

	// Run before every child command executes run
	PersistentPreRun: func (cmd *cobra.Command, args []string) {

		// Warn if blackfire environment not setup
		if blackFire := viper.GetString("BLACKFIRE_SERVER_ID"); len(blackFire) <= 0 {
			c := color.New(color.FgHiYellow).Add(color.Bold)
			c.Println("Warning: You do not have blackfire credentials stored in your environment! You will not be able to use blackfire until you add them!")
			c.Println("To add them easily, copy the export lines from https://blackfire.io/docs/integrations/docker (server are likely all you'll use) and add them to your ~/.zshrc \n")
		}

		// If minikube is not running and local environment is selected, ask user if they'd like us to start it
		if out, _ := util.ExecCmdChain("minikube status | grep 'localkube: Running'"); len(out) <= 0 && environment == "local" {
			if approve := util.GetApproval("Minikube is not running, would you like to start it?"); approve {
				components.StartCmd.Run(cmd, args)
			} else {
				color.Red("You need to start minikube before deploying a project!")
				os.Exit(1)
			}
		}
	},
}

var environment string
func init() {
	ProjectCmd.AddCommand(createCmd)
	ProjectCmd.AddCommand(syncCmd)
	ProjectCmd.AddCommand(deleteCmd)
	ProjectCmd.AddCommand(setCmd)
	ProjectCmd.AddCommand(updateCmd)
	ProjectCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "local", "Define which environment to apply argo commands to. Ex: \"local\", \"dev\", or \"prod\".")
	viper.BindPFlag("environment", ProjectCmd.PersistentFlags().Lookup("environment"))
}

// Shared functions

// Attempt to create or find project
func locateProject(args []string) (string, error) {
	if hardSetName := viper.GetString("project-name"); len(hardSetName) > 0 {
		return hardSetName, nil
	}
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
		// Check for presence of argoyml
		noConfig := viper.GetBool("noConfig")
		if noConfig {
			return projectName, errors.New("No argo configuration file found!  Please re-run this command in a project with an argo configuration file in it's root, or specifiy a git repo to clone.")
		} else {
			cwdPath, _ := os.Getwd()
			projectName = filepath.Base(cwdPath)
			color.Cyan("Reading project %s from argo config file...", projectName)
		}
	}
	return projectName, err
}

// Update context/project/etc to match environment
func setupKubectl(name string, environment string, set bool) {
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

	approve := set || util.GetApproval("Gcloud configuration has been updated, would you like to create and switch to a new kubectl context?");
	if approve {
		// Setup a kubectl context and switch to it
		util.ExecCmd("kubectl", "config", "delete-context", name)
		util.ExecCmd("kubectl", "config", "set-context", name, fmt.Sprintf("--cluster=%s", contextCluster), fmt.Sprintf("--user=%s", contextUser), fmt.Sprintf("--namespace=%s", name))
		util.ExecCmd("kubectl", "config", "use-context", name)
		color.Cyan("Created new %s kubectl context and set to active.", name)
	}
}

// Accept projectname and chart to install project infra via helm
func helmUpgrade(projectName string) error {

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

		// Push using latest tag or CIRCLE_SHA1 if running in circle environment/is otherwise provided.
		if circleSha := viper.GetString("CIRCLE_SHA1"); len(circleSha) > 0 {
			appImage = fmt.Sprintf("%s:%s", appImage, circleSha)
		} else {
			appImage = fmt.Sprintf("%s:%s", appImage, "latest")
		}

		helmValues = append(helmValues, fmt.Sprintf("application.image=%s", appImage))
	}

	command := fmt.Sprintf("helm upgrade --debug --install %s %s --set %s", projectName, viper.GetString("chart"), strings.Join(helmValues, ","))
	out, err := util.ExecCmdChainCombinedOut(command)
	if (err != nil) {
		color.Red(out)
	} else if debugMode := viper.GetString("debug"); len(debugMode) > 0 {
		color.Green(out)
	}
	return err
}