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
	"regexp"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage individual argo projects.",
	Long: `Manipulate argo projects.`,

	// Run before every child command executes run
	PersistentPreRun: func (cmd *cobra.Command, args []string) {
		initProjectConfig()

		chart := projectConfig.GetString("chart");
		oldChart, _ := regexp.MatchString("drupal-8-1", chart)
		if (oldChart) {
			color.Red("This chart's version is now unsupported.  Either upgrade the project to >2.0.0 or switch to an argo version below 1.0.0")
			os.Exit(1)
		}

		projectName := projectConfig.GetString("project_name");
		if len(projectName) == 0 {
			color.Red("You need to specify a project name in argo.yml!")
			os.Exit(1)
		}

		// If minikube is not running and local environment is selected, ask user if they'd like us to start it
		if projectConfig.GetString("environment") == "local" {
			if out, _ := util.ExecCmdChain("minikube status | grep 'minikube: Running'"); len(out) <= 0  {
				if approve := util.GetApproval("Minikube is not running, would you like to start it?"); approve {
					components.StartCmd.Run(cmd, args)
				} else {
					color.Red("You need to start minikube before deploying a project!")
					os.Exit(1)
				}
			}
		}

		// Warn if blackfire environment not setup
		if blackFire := projectConfig.GetString("BLACKFIRE_SERVER_ID"); len(blackFire) <= 0 {
			c := color.New(color.FgHiYellow).Add(color.Bold)
			c.Println("Warning: You do not have blackfire credentials stored in your environment! You will not be able to use blackfire until you add them!")
			c.Println("To add them easily, copy the export lines from https://blackfire.io/docs/integrations/docker (server are likely all you'll use) and add them to your ~/.zshrc \n")
		}
	},
}

// Values for these are loaded in initProjectConfig()
// Entire package will use projectConfig viper instance
var projectConfig = viper.New()
// Initialize a new config for environment specific variables
var envConfig = viper.New()

func init() {
	ProjectCmd.PersistentFlags().StringP("environment", "e", "", "Define which environment to apply argo commands.  This is generally the git branch.")
	projectConfig.BindPFlag("environment", ProjectCmd.PersistentFlags().Lookup("environment"))

	ProjectCmd.PersistentFlags().Bool("wait", false, "If true, apply --wait to helm commands.  See helm documentation for more details.")
	projectConfig.BindPFlag("wait", ProjectCmd.PersistentFlags().Lookup("wait"))

	ProjectCmd.AddCommand(createCmd)
	ProjectCmd.AddCommand(syncCmd)
	ProjectCmd.AddCommand(deleteCmd)
	ProjectCmd.AddCommand(setEnvCmd)
	ProjectCmd.AddCommand(updateCmd)
	ProjectCmd.AddCommand(rollbackCmd)
	ProjectCmd.AddCommand(existsCmd)
}

func initProjectConfig() {
	projectConfig.SetConfigName("config") 	// Name of config file (without extension)
	projectConfig.AddConfigPath("./.argo")  // Current directory/.argo
	// TODO - Perhaps only provide access to environment variables in a global viper object - MEA
	projectConfig.AutomaticEnv()

	// Error if no yaml found
	if err := projectConfig.ReadInConfig(); err != nil {
		color.Red("%s",err)
		color.Red("Error locating or parsing .argo/config.yml!")
		os.Exit(1)
	}

	// The sub-command `sync` does not use an --environment flag.
	// TODO - Ditch the --environment flag - MEA
	if (len(projectConfig.GetString("environment")) > 0) {
		envConfig = setupEnvConfigViper(projectConfig.GetString("environment"))
	} else {
		if viper.GetBool("debug") {
			color.Yellow("[debug] - Skipping environment configuration import from -e/--environment flag, command does not use")
		}
	}
}

// Grab the configuration for the intended environment.
// Every environment should have an associated config yaml in the project's .argo/environments folder
func setupEnvConfigViper(environmentFlag string) *viper.Viper {
	config := viper.New()
	config.SetConfigName(environmentFlag)
	config.AddConfigPath("./.argo/environments")
	if err := config.ReadInConfig(); err != nil {
		color.Red("%s",err)
		color.Red("Error locating or parsing %s env's helm value file (should be ./argo/environments/%s.yaml!", environmentFlag)
		os.Exit(1)
	}
	return config
}

// Setup values to use for gcloud and kubectl commands for specified environment
func setKubectlConfig(environment string) {
	currentEnvConfig := setupEnvConfigViper(environment)
	var namespace = currentEnvConfig.GetString("general.namespace")

	var contextCluster string
	if environment == "local" {
		contextCluster = "minikube"
	} else {
		gcloudCluster := currentEnvConfig.GetString("gcp.cluster")
		gcloudProject := currentEnvConfig.GetString("gcp.project")
		gcloudZone := currentEnvConfig.GetString("gcp.compute_zone")

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

	color.Cyan("Recreating kubectl context and setting to active...")

	contextName := namespace
	// Setup a kubectl context and switch to it
	util.ExecCmd("kubectl", "config", "delete-context", contextName)
	util.ExecCmd("kubectl", "config", "set-context", contextName, fmt.Sprintf("--cluster=%s", contextCluster), fmt.Sprintf("--user=%s", contextCluster), fmt.Sprintf("--namespace=%s", namespace))
	util.ExecCmd("kubectl", "config", "use-context", contextName)
	color.Cyan("Created new %s kubectl context and set to active.", contextName)
}

// Run a check to see if the project already exists in helm
// TODO - this is not necessarily switching to the correct kubectl context before running resulting in false negatives
func checkExisting() bool {
	projectName := projectConfig.GetString("project_name")
	environment := projectConfig.GetString("environment")

	color.Cyan("Ensuring existing helm project does not already exist...")
	out, err := util.ExecCmdChainCombinedOut(fmt.Sprintf("helm status %s-%s", projectName, environment))

	if (err != nil) {
		color.Red(out)
		return false
	} else {
		return true
	}
}

// Construct values passed directly into helm via --set
// Other values are passed through the environment values.yaml
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

// Helm upgrade is run on deploy and update
func helmUpgrade() error {
	projectName := projectConfig.GetString("project_name")
	environment := projectConfig.GetString("environment")

	color.Cyan("Installing project chart via helm...")

	var waitFlag string
	if projectConfig.GetBool("wait") {
		// Intended for CI, wait for updated infrastructure to apply fully so subsequent commands (drush) run against new infra
		waitFlag = "--wait"
		color.Cyan("Using wait, command will take a moment...")
	}

	var helmValues HelmValues

	helmValues.appendValue("general.project_name", projectName, true);
	helmValues.appendValue("general.environment", environment, true);

	// These come from environment vars
	// TODO - Blackfire credential management? Currently deploying to both environments - MEA
	helmValues.appendProjectValue("applications.blackfire.server_id", "BLACKFIRE_SERVER_ID", false)
	helmValues.appendProjectValue("applications.blackfire.server_token", "BLACKFIRE_SERVER_TOKEN", false)
	helmValues.appendProjectValue("applications.blackfire.client_id", "BLACKFIRE_CLIENT_ID", false)
	helmValues.appendProjectValue("applications.blackfire.client_token", "BLACKFIRE_CLIENT_TOKEN", false)

	// Derived values
	if environment == "local" {
		// Using https://github.com/mstrzele/minikube-nfs
		// Creates a persistent volume with contents of /Users mounted from an nfs share from host
		// So rather than using /Users directly, grab the path within the project
		// Check the helm chart mounts
		projectPath, _ := util.ExecCmdChain("printf ${PWD#/Users/}")
		helmValues.appendValue("applications.drupal.local.project_root", projectPath, true)
		localIp, _ := util.ExecCmdChain("ifconfig | grep \"inet \" | grep -v 127.0.0.1 | awk '{print $2}' | sed -n 1p")
		helmValues.appendValue("applications.xdebug.host_ip", localIp, true)
	} else {
		// Obtain the git commit from env vars if present in CircleCI
		// TODO - Make this a required argument rather than inferred environment variable
		if circleSha := projectConfig.GetString("CIRCLE_SHA1"); len(circleSha) > 0 {
			helmValues.appendValue("general.git_commit", circleSha, false)
		} else {
			if approve := util.GetApproval("No CIRCLE_SHA1 environment variable set (should be git sha1 hash of commit), use 'latest' tag for application container?"); !approve {
				color.Red("User exited.  Please run 'export CIRCLE_SHA1=' adding your targetted git commit hash, or run again and agree to use 'latest.'")
				os.Exit(1)
			}
		}
	}

	chartConfigPath := fmt.Sprintf("charts.%s", environment)
	helmChartPath := projectConfig.GetString(chartConfigPath)

	command := fmt.Sprintf("helm upgrade --install --values %s %s %s-%s %s --set %s",
		envConfig.ConfigFileUsed(),
		waitFlag,
		projectName,
		environment,
		helmChartPath,
		strings.Join(helmValues.values, ","))
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

