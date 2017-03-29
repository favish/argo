package project

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/fatih/color"
	"github.com/favish/argo/util"
	"github.com/favish/argo/cmd/components"
	"os"
	"strings"
	"fmt"
	"path"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage individual argo projects.",
	Long: `Manipulate argo projects.`,

	// Run before every child command executes run
	PersistentPreRun: func (cmd *cobra.Command, args []string) {
		initProjectConfig()

		projectName := projectConfig.GetString("project-name");
		if len(projectName) == 0 {
			color.Red("You need to specify a project name in argo.yml!")
			os.Exit(1)
		}

		// If minikube is not running and local environment is selected, ask user if they'd like us to start it
		if out, _ := util.ExecCmdChain("minikube status | grep 'localkube: Running'"); len(out) <= 0 && projectConfig.GetString("environment") == "local" {
			if approve := util.GetApproval("Minikube is not running, would you like to start it?"); approve {
				components.StartCmd.Run(cmd, args)
			} else {
				color.Red("You need to start minikube before deploying a project!")
				os.Exit(1)
			}
		}

		setKubectlConfig(projectConfig.GetString("environment"))

		// Warn if blackfire environment not setup
		if blackFire := projectConfig.GetString("BLACKFIRE_SERVER_ID"); len(blackFire) <= 0 {
			c := color.New(color.FgHiYellow).Add(color.Bold)
			c.Println("Warning: You do not have blackfire credentials stored in your environment! You will not be able to use blackfire until you add them!")
			c.Println("To add them easily, copy the export lines from https://blackfire.io/docs/integrations/docker (server are likely all you'll use) and add them to your ~/.zshrc \n")
		}
	},
}

// Entire package will use projectConfig viper instance
var projectConfig = viper.New()

func init() {
	ProjectCmd.PersistentFlags().StringP("environment", "e", "local", "Define which environment to apply argo commands to. Ex: \"local\", \"dev\", or \"prod\".")
	projectConfig.BindPFlag("environment", ProjectCmd.PersistentFlags().Lookup("environment"))

	ProjectCmd.PersistentFlags().Bool("wait", false, "If true, apply --wait to helm commands.  See helm documentation for more details.")
	projectConfig.BindPFlag("wait", ProjectCmd.PersistentFlags().Lookup("wait"))

	ProjectCmd.AddCommand(createCmd)
	ProjectCmd.AddCommand(syncCmd)
	ProjectCmd.AddCommand(deleteCmd)
	ProjectCmd.AddCommand(setEnvCmd)
	ProjectCmd.AddCommand(updateCmd)
}

func initProjectConfig() {
	projectConfig.SetConfigName("argo") 	// Name of config file (without extension)
	projectConfig.AddConfigPath(".")  	// Current directory
	// TODO - Perhaps only provide access to environment variables in a global viper object - MEA
	projectConfig.AutomaticEnv()

	// Error if no yaml found
	if err := projectConfig.ReadInConfig(); err != nil {
		color.Red("%s",err)
		color.Red("No project argo.yml found in this directory!")
		os.Exit(1)
	}
}


// Setup values to use for gcloud and kubectl commands for specified environment
func setKubectlConfig(environment string) {
	projectName := projectConfig.GetString("project-name");

	var contextCluster string
	if environment == "local" {
		contextCluster = "minikube"
	} else {
		gcloudProject := viper.GetString(fmt.Sprintf("environments.%s.project", environment))
		gcloudZone := viper.GetString(fmt.Sprintf("environments.%s.compute-zone", environment))
		gcloudCluster := viper.GetString(fmt.Sprintf("environments.%s.cluster", environment))

		gcloudCmd := fmt.Sprintf("container clusters get-credentials %s --project=%s --zone=%s", gcloudCluster, gcloudProject, gcloudZone)

		// To use argo as a deployment tool in CircleCI, gcloud has to be invoked as sudo with explicit binary path
		// CircleCI 2.0 does not have this stipulation, identified by presence of CIRCLE_STAGE var
		if projectConfig.GetString("CIRCLECI") == "true" && len(projectConfig.GetString("CIRCLE_STAGE")) == 0 {
			gcloudCmd = fmt.Sprintf("sudo /opt/google-cloud-sdk/bin/gcloud %s", gcloudCmd)
		} else {
			gcloudCmd = fmt.Sprintf("gcloud %s", gcloudCmd)
		}

		if _, err := util.ExecCmdChain(gcloudCmd); err != nil {
			color.Red("Error getting kubectl cluster credentials via gcloud! %s", err)
			os.Exit(1)
		}

		contextCluster = fmt.Sprintf("gke_%s_%s_%s", gcloudProject, gcloudZone, gcloudCluster)
	}

	// If the namespace does not exist, create one.
	if err := util.ExecCmd("kubectl", "get", "namespace", projectName); err != nil {
		util.ExecCmd("kubectl", "create", "namespace", projectName)
		color.Cyan("Created new %s kubernetes namespace.", projectName)
	}

	color.Cyan("Recreating kubectl context and setting to active...")

	contextName := fmt.Sprintf("%s-%s", projectName, environment)
	// Setup a kubectl context and switch to it
	util.ExecCmd("kubectl", "config", "delete-context", contextName)
	util.ExecCmd("kubectl", "config", "set-context", contextName, fmt.Sprintf("--cluster=%s", contextCluster), fmt.Sprintf("--user=%s", contextCluster), fmt.Sprintf("--namespace=%s", projectName))
	util.ExecCmd("kubectl", "config", "use-context", contextName)
	color.Cyan("Created new %s kubectl context and set to active.", contextName)

}

// Run a check to see if the project already exists in helm
func checkExisting() bool {
	projectName := projectConfig.GetString("project-name")

	color.Cyan("Ensuring existing helm project does not already exist...")
	projectExists := false
	if out, _ := util.ExecCmdChain(fmt.Sprintf("helm status %s | grep 'STATUS: DEPLOYED'", projectName)); len(out) > 0 {
		color.Yellow(out)
		projectExists = true
	}
	return projectExists
}

// Accept projectname and chart to install project infra via helm
func helmUpgrade() error {
	projectName := projectConfig.GetString("project-name")
	environment := projectConfig.GetString("environment")

	color.Cyan("Installing project chart via helm...")

	var waitFlag string
	if projectConfig.GetBool("wait") {
		waitFlag = "--wait"
		color.Cyan("Using wait, command will take a moment...")
	}

	var helmValues []string

	helmValues = append(helmValues, fmt.Sprintf("namespace=%s", projectName))
	helmValues = append(helmValues, fmt.Sprintf("environment_type=%s", environment))

	// TODO - Blackfire credential management? Currently deploying to both environments - MEA
	helmValues = append(helmValues, fmt.Sprintf("blackfire.server_id=%s", projectConfig.GetString("BLACKFIRE_SERVER_ID")))
	helmValues = append(helmValues, fmt.Sprintf("blackfire.server_token=%s", projectConfig.GetString("BLACKFIRE_SERVER_TOKEN")))

	helmValues = append(helmValues, fmt.Sprintf("php_image=%s", projectConfig.GetString("php-image")))
	helmValues = append(helmValues, fmt.Sprintf("php_xdebug_image=%s", projectConfig.GetString("php-xdebug-image")))
	helmValues = append(helmValues, fmt.Sprintf("nginx_image=%s", projectConfig.GetString("nginx-image")))
	helmValues = append(helmValues, fmt.Sprintf("web_image=%s", projectConfig.GetString("web-image")))

	helmValues = append(helmValues, fmt.Sprintf("application.env=%s", environment))

	if environment == "local" {
		helmValues = append(helmValues, fmt.Sprintf("local.webroot=%s", path.Join(projectConfig.GetString("PWD"), projectConfig.GetString("environments.local.webroot"))))
		helmValues = append(helmValues, fmt.Sprintf("local.project_root=%s", projectConfig.GetString("PWD")))
		helmValues = append(helmValues, fmt.Sprintf("mysql.db=%s", projectConfig.GetString("environments.local.mysql.db")))
		helmValues = append(helmValues, fmt.Sprintf("mysql.pass=%s", projectConfig.GetString("environments.local.mysql.pass")))
		helmValues = append(helmValues, fmt.Sprintf("mysql.user=%s", projectConfig.GetString("environments.local.mysql.user")))
		localIp, _ := util.ExecCmdChain("ifconfig | grep \"inet \" | grep -v 127.0.0.1 | awk '{print $2}' | sed -n 1p")
		helmValues = append(helmValues, fmt.Sprintf("local.host_ip=%s", localIp))
	} else {
		project := projectConfig.GetString(fmt.Sprintf("environments.%s.project", environment))
		computeZone := projectConfig.GetString(fmt.Sprintf("environments.%s.compute-zone", environment))
		instance := projectConfig.GetString(fmt.Sprintf("environments.%s.mysql.instance", environment))

		database := projectConfig.GetString(fmt.Sprintf("environments.%s.mysql.db", environment))

		mysqlInstance := fmt.Sprintf("%s:%s:%s", project, computeZone, instance)

		helmValues = append(helmValues, fmt.Sprintf("mysql.instance=%s", mysqlInstance))
		helmValues = append(helmValues, fmt.Sprintf("mysql.db=%s", database))

		appImage := projectConfig.GetString(fmt.Sprintf("environments.%s.application-image", environment))
		// Push using latest tag or CIRCLE_SHA1 if running in circle environment/is otherwise provided.
		if circleSha := projectConfig.GetString("CIRCLE_SHA1"); len(circleSha) > 0 {
			appImage = fmt.Sprintf("%s:%s", appImage, circleSha)
		} else {
			appImage = fmt.Sprintf("%s:%s", appImage, "latest")
		}

		helmValues = append(helmValues, fmt.Sprintf("application.image=%s", appImage))
	}

	if environment == "dev" {
		helmValues = append(helmValues, fmt.Sprintf("dev.basic_auth=%s", projectConfig.GetString("environments.dev.basic-auth")))
	}

	command := fmt.Sprintf("helm upgrade --install %s %s %s --set %s", waitFlag, projectName, projectConfig.GetString("chart"), strings.Join(helmValues, ","))
	out, err := util.ExecCmdChainCombinedOut(command)
	if (err != nil) {
		color.Red(out)
		os.Exit(1)
	} else if debugMode := viper.GetString("debug"); len(debugMode) > 0 {
		color.Green(out)
	}
	return err
}