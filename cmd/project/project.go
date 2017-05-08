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
		if projectConfig.GetString("environment") == "local" {
			if out, _ := util.ExecCmdChain("minikube status | grep 'localkube: Running'"); len(out) <= 0  {
				if approve := util.GetApproval("Minikube is not running, would you like to start it?"); approve {
					components.StartCmd.Run(cmd, args)
				} else {
					color.Red("You need to start minikube before deploying a project!")
					os.Exit(1)
				}
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
	ProjectCmd.AddCommand(rollbackCmd)
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

type HelmValues struct {
	values []string
}
// Add value directly with resolved string
func (hv *HelmValues) appendValue(helmKey string, value string, required bool) {
	if (len(value) <= 0) {
		if (required) {
			color.Red("Missing required helm value for %s", helmKey)
			os.Exit(1)
		} else if (viper.GetBool("debug")) {
			color.Yellow("[debug] - Not setting value for helm key %s", helmKey)
		}
	} else {
		if (viper.GetBool("debug")) {
			color.Yellow("[debug] - Setting helm key %s to %s", helmKey, value)
		}
		hv.values = append(hv.values, fmt.Sprintf("%s=%s", helmKey, value))
	}
}

// Add value from argo.yml or other config value (environment variables) by key
func (hv *HelmValues) appendProjectValue(helmKey string, projectKey string, required bool) {
	projectValue := projectConfig.GetString(projectKey)
	hv.appendValue(helmKey, projectValue, required)
}

// Grab project value based on environment value
// All environment values live at environments.%s, ie environments.dev or environments.local
func (hv *HelmValues) appendProjectEnvValue(helmKey string, projectKey string, environment string, required bool) {
	environmentValue := projectConfig.GetString(fmt.Sprintf("environments.%s.%s", environment, projectKey))
	hv.appendValue(helmKey, environmentValue, required)
}

func helmUpgrade() error {
	projectName := projectConfig.GetString("project-name")
	environment := projectConfig.GetString("environment")

	color.Cyan("Installing project chart via helm...")

	var waitFlag string
	if projectConfig.GetBool("wait") {
		// Intended for CI, wait for updated infrastructure to apply fully so subsquent commands (drush) run against new infra
		waitFlag = "--wait"
		color.Cyan("Using wait, command will take a moment...")
	}

	var helmValues HelmValues

	helmValues.appendValue("namespace", projectName, true)
	helmValues.appendValue("environment_type", environment, true)
	helmValues.appendValue("application.env", environment, true)

	helmValues.appendProjectEnvValue("hostname", "hostname", environment, false)
	helmValues.appendProjectEnvValue("redirect_www", "redirect-www", environment, false)

	// TODO - Blackfire credential management? Currently deploying to both environments - MEA
	helmValues.appendProjectValue("blackfire.server_id", "BLACKFIRE_SERVER_ID", false)
	helmValues.appendProjectValue("blackfire.server_token", "BLACKFIRE_SERVER_TOKEN", false)
	helmValues.appendProjectValue("blackfire.client_id", "BLACKFIRE_CLIENT_ID", false)
	helmValues.appendProjectValue("blackfire.client_token", "BLACKFIRE_CLIENT_TOKEN", false)

	if environment == "local" {
		helmValues.appendProjectValue("local.project_root", "PWD", true)
		helmValues.appendProjectValue("local.theme_dir", "environments.local.theme_dir", true)
		helmValues.appendProjectValue("mysql.db", "environments.local.mysql.db", true)
		helmValues.appendProjectValue("mysql.pass", "environments.local.mysql.pass", true)
		helmValues.appendProjectValue("mysql.user", "environments.local.mysql.user", true)

		localIp, _ := util.ExecCmdChain("ifconfig | grep \"inet \" | grep -v 127.0.0.1 | awk '{print $2}' | sed -n 1p")
		helmValues.appendValue("local.host_ip=%s", localIp, true)
	} else {
		cidrConfigValue := projectConfig.GetString(fmt.Sprintf("environments.%s.container-cidr", environment))
		helmValues.appendValue("container_cidr", cidrConfigValue, true)

		helmValues.appendProjectEnvValue("mysql.instance", "mysql.instance", environment, true)

		helmValues.appendProjectEnvValue("mysql.db", "mysql.db", environment, true)

		helmValues.appendProjectEnvValue("cron.host", "cron.host", environment, false)
		helmValues.appendProjectEnvValue("cron.key", "cron.key", environment, false)

		appImage := projectConfig.GetString(fmt.Sprintf("environments.%s.application-image", environment))
		// Push using latest tag or CIRCLE_SHA1 if running in circle environment/is otherwise provided.
		if circleSha := projectConfig.GetString("CIRCLE_SHA1"); len(circleSha) > 0 {
			appImage = fmt.Sprintf("%s:%s", appImage, circleSha)
		} else {
			appImage = fmt.Sprintf("%s:%s", appImage, "latest")
		}
		helmValues.appendValue("application.image", appImage, true)

		helmValues.appendProjectEnvValue("web_node_port", "web-node-port", environment, true)
	}

	if environment == "dev" {
		helmValues.appendProjectValue("dev.basic_auth", "environments.dev.basic-auth", true)
		helmValues.appendProjectValue("dev.basic_auth_node_port", "environments.dev.basic-auth-node-port", true)
	}

	command := fmt.Sprintf("helm upgrade --install %s %s %s --set %s", waitFlag, projectName, projectConfig.GetString("chart"), strings.Join(helmValues.values, ","))
	out, err := util.ExecCmdChainCombinedOut(command)
	if (err != nil) {
		color.Red(out)
		if (projectConfig.GetBool("rollback-on-failure")) {
			color.Yellow("Your helm upgrade resulted in a failure, attempting to rollback...")
			rollbackCmd.Run(nil, nil)
			color.Yellow("Successfully rolled back attempted update, exiting with error.  You will want to correct this.")
			os.Exit(1)
		} else {
			os.Exit(1)
		}
	} else if debugMode := viper.GetString("debug"); len(debugMode) > 0 {
		color.Green(out)
	}
	return err
}

