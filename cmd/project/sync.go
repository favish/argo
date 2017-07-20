package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
	"os"
	"fmt"
	"strings"
	"time"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update local copies of persistent data to match remote dev.",
	Long: `
		Sync local copies of remote persistent data (user files/database) to local argo environment.
	`,
}

var filesCmd = &cobra.Command{
	Use: "files",
	Short: "Sync files SOURCE > DESTINATION.",
	Long: `Uses temporary containers to perform an rsync between two separate environments`,
	Run: func(cmd *cobra.Command, args []string) {
		validateArgs(args)

		srcEnv := args[0]
		destEnv := args[1]

		if approve := util.GetApproval(fmt.Sprintf("This will sync files from %s to %s, proceed?", srcEnv, destEnv)); !approve {
			color.Red("Skipping sync.")
			os.Exit(1)
		} else {
			helmChart := "/Users/mikaAguilar/Projects/helm-charts/rsync"

			srcConfig := setupEnvConfigViper(srcEnv)
			destConfig := setupEnvConfigViper(destEnv)

			timestamp := int32(time.Now().Unix())

			color.Cyan("Generating a keypair to use for this transaction...")
			util.ExecCmdChain(fmt.Sprintf("ssh-keygen -q -b 2048 -t rsa -N '' -f /tmp/argo-temp-%v", timestamp))
			privateKey, _ := util.ExecCmdChain(fmt.Sprintf("cat /tmp/argo-temp-%v | base64 | tr -d '\n'", timestamp))
			publicKey, _ := util.ExecCmdChain(fmt.Sprintf("cat /tmp/argo-temp-%v.pub | base64 | tr -d '\n'", timestamp))
			util.ExecCmd("rm", fmt.Sprintf("/tmp/argo-temp-%v*", timestamp))

			color.Cyan("Enabling temporary access to source...")

			projectName := projectConfig.GetString("project_name")
			deploymentName := fmt.Sprintf("rsync-%s-%v", projectName, timestamp)

			var srcHelmValues HelmValues
			// Check if source and dest exist on different clusters, will also detect local
			// First, concat project and cluster, in case clusters are named the same
			srcClusterLocation := fmt.Sprintf("%s-%s", srcConfig.GetString("gcp.project"), srcConfig.GetString("gcp.cluster"))
			destClusterLocation := fmt.Sprintf("%s-%s", destConfig.GetString("gcp.project"), destConfig.GetString("gcp.cluster"))
			externalCluster := false
			if (srcClusterLocation != destClusterLocation) {
				externalCluster = true
				srcHelmValues.appendValue("service_type", "LoadBalancer", true)
				color.Cyan("Clusters do not match, source will expose endpoint via LoadBalancer")
			} else {
				color.Cyan("Same cluster sync, using ClusterIP")
			}

			srcHelmValues.appendValue("name", deploymentName, true)
			srcHelmValues.appendValue("namespace", srcConfig.GetString("namespace"), true)
			srcHelmValues.appendValue("volume", fmt.Sprintf("%s-default-files-nfs", srcConfig.GetString("namespace")), true)
			srcHelmValues.appendValue("is_source", "true", true)
			srcHelmValues.appendValue("private_key", privateKey, true)
			srcHelmValues.appendValue("public_key", publicKey, true)
			// Install rsync chart on source cluster, get the service or pod ip and leave it open so destination can pull from it
			color.Cyan("Spinning up pod on source to make source volume available...")
			setKubectlConfig(srcEnv)
			out, err := util.ExecCmdChainCombinedOut(fmt.Sprintf("helm install --wait --name %s %s --set %s",
				deploymentName,
				helmChart,
				strings.Join(srcHelmValues.values, ","),
			))
			if (err != nil) {
				color.Red("Error installing on source: %s", out)
				os.Exit(1)
			} else {
				color.Cyan(out)
			}

			var sourceIP string
			if (externalCluster) {
				sourceIP, err = util.ExecCmdChain(fmt.Sprintf("kubectl get service %s -o jsonpath='{.status.loadBalancer.ingress[0].ip}'", deploymentName))
			} else {
				sourceIP, err = util.ExecCmdChain(fmt.Sprintf("kubectl get service %s -o jsonpath='{.spec.clusterIP}'", deploymentName))
			}
			if (err != nil) {
				color.Red("Unable to get sourceIP, rolling back. (%s)", err)
				setKubectlConfig(srcEnv)
				util.ExecCmd("helm", "delete", "--purge", deploymentName)
				os.Exit(1)
			}

			// Spin up chart on destination, which starts an rsync job
			var destHelmValues HelmValues
			destDeploymentName := fmt.Sprintf("%s-dest", deploymentName)
			destHelmValues.appendValue("name", destDeploymentName, true)
			destHelmValues.appendValue("namespace", destConfig.GetString("namespace"), true)
			destHelmValues.appendValue("is_destination", "true", true)
			destHelmValues.appendValue("volume", fmt.Sprintf("%s-default-files-nfs", destConfig.GetString("namespace")), true)
			destHelmValues.appendValue("private_key", privateKey, true)
			destHelmValues.appendValue("public_key", publicKey, true)
			color.Cyan("Starting destination rsync process...")
			setKubectlConfig(destEnv)
			out, err = util.ExecCmdChainCombinedOut(fmt.Sprintf("helm install --wait --name %s %s --set %s",
				destDeploymentName,
				helmChart,
				strings.Join(destHelmValues.values, ","),
			))

			if (err != nil) {
				color.Red("Error installing on destination: %s", out)
				os.Exit(1)
			} else {
				color.Cyan(out)
			}

			util.ExecCmd("kubectl", "get", "pods")
			// Get pod name on dest, execute rsync on it
			podName, err := util.ExecCmdChainCombinedOut("kubectl get pods --selector='task=rsync' | grep 'Running' | awk '{print $1}' | tr -d '\n'")
			if (err != nil || len(podName) == 0) {
				color.Red("Error obtaining pod: %s", out)
			}

			out, err = util.ExecCmdChainCombinedOut(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- rsync -avz -e 'ssh -o StrictHostKeyChecking=no' %s:/srv/ /srv", podName, sourceIP))
			if (err != nil) {
				color.Red("Error executing rsync: %s", out)
			} else {
				color.Cyan(out)
			}

			util.ExecCmd("kubectl", "get", "pods")
		 	// Cleanup
			color.Cyan("Cleaning up deployments...")
			util.ExecCmd("helm", "delete", "--purge", destDeploymentName)
			setKubectlConfig(srcEnv)
			util.ExecCmd("helm", "delete", "--purge", deploymentName)

		}
	},
}

var dbCmd = &cobra.Command{
	Use: "db",
	Short: "Sync database SOURCE > DESTINATION.  Requires functional database on both targets (run drush site-install if needed).",
	Long: `Use this command to retrieve or push your database to and from remote environments.
	Ex. "argo project sync db dev local" brings dev db to your machine.
	Light wrapper around drush, implementation similar to drush sql-sync in a world without ssh.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		validateArgs(args)

		if (args[0] == "prod" || args[1] == "prod") {
			color.Red("WARNING: This Command performs a sqldump, which causes the database to READ LOCK during the dump operation.")
			color.Red("WARNING: During this time, the sql database will be effectively inaccessible to applications, and the site will go down.")
		}
		if approve := util.GetApproval("Syncing databases will cause a temporary service outage for both targets, are you sure?"); !approve {
			return
		}

		source := args[0]
		destination := args[1]

		// Tag dumps with current timestamp
		timestamp := int32(time.Now().Unix())
		databaseFilename := fmt.Sprintf("argo-db-tmp-%d.sql", timestamp)

		// Process:
		// Connect to SOURCE application pod
		// Dump database to temp file
		// Copy database to operator's machine
		// Connect to DESTINATION application pod
		// Make sure database exists on DESTINATION
		// Create temp backup of DESTINATION database
		// Drop DESTINATION application database
		// Copy dump to DESTINATION pod, unzip there
		// Sync DESTINATION application database
		// Delete operator's temp file

		// Point kubectl at FROM, dump database to local temp file
		setKubectlConfig(source)
		// Get the container running application
		sourcePodName, err := util.ExecCmdChain("kubectl get pods --selector='service=cloud-command' | grep 'Running' | awk '{print $1}' | tr -d '\n'")
		if (err != nil) {
			color.Red("Error getting pod name: %s", err)
			os.Exit(1)
		}

		// Dump database on FROM environment
		color.Cyan("Creating a database dump on FROM container...")
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- drush sql-dump --gzip --result-file=/tmp/%s",
			sourcePodName,
			databaseFilename))
		if (err != nil) {
			color.Red("Error creating dump: %s", err)
			os.Exit(1)
		}

		// Copy database to operator's machine
		color.Cyan("Copying dump from FROM container to your machine...")
		err = util.ExecCmd("kubectl", "cp",
			fmt.Sprintf("%s:/tmp/%s.gz", sourcePodName, databaseFilename),
			fmt.Sprintf("/tmp/%s.gz", databaseFilename))
		if (err != nil) {
			color.Red("Error copying dump locally: %s", err)
			os.Exit(1)
		}

		// Point kubectl at DEST
		color.Cyan("Redirecting kubectl config to %s environment...", destination)
		setKubectlConfig(destination)

		// Get the container running application on destination
		destPodName, err := util.ExecCmdChain("kubectl get pods --selector='service=cloud-command' | grep 'Running' | awk '{print $1}' | tr -d '\n'")
		if (err != nil) {
			color.Red("Error getting pod name: %s", err)
			os.Exit(1)
		}

		color.Cyan("Ensuring database exists on %s...", destination)
		util.ExecCmdChain(fmt.Sprintf("kubectl exec %s /opt/init_database.sh", destPodName))

		color.Cyan("Creating backup of %s before dropping existing database, saving to /var/www/%s.bak.gz", destination, databaseFilename)
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- /bin/bash -c 'drush sql-dump --gzip --result-file=/tmp/argo-db-tmp.sql.bak'", destPodName))
		if (err != nil) {
			color.Red("Error backing up: %s", err)
			os.Exit(1)
		}

		color.Cyan("Dropping %s database...", destination)
		if approve := util.GetApproval("YOU ARE NOW DROPPING THE DESTINATION DATABASE TO REPLACE WITH DUMP, ARE YOU SURE?"); !approve {
			return
		}
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- /bin/bash -c 'drush sql-drop -y'", destPodName))

		color.Cyan("Uploading database to %s via kubectl cp...", destination)
		err = util.ExecCmd("kubectl",
			"cp",
			fmt.Sprintf("/tmp/%s.gz", databaseFilename),
			fmt.Sprintf("%s:/tmp/%s.gz", destPodName, databaseFilename))
		if (err != nil) {
			color.Red("Error uploading database dump: %s", err)
			os.Exit(1)
		}

		color.Cyan("Unzipping...")
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- gunzip /tmp/%s.gz", destPodName, databaseFilename))
		if (err != nil) {
			color.Red("Error unzipping database dump: %s", err)
			os.Exit(1)
		}

		color.Cyan("Importing database dump to %s environment, may take a few moments...", destination)
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- /bin/bash -c 'drush sqlc < /tmp/%s'", destPodName, databaseFilename))
		if (err != nil) {
			color.Red("Error importing database: %s", err)
			os.Exit(1)
		}

		err = util.ExecCmd("rm", fmt.Sprintf("/tmp/%s.gz", databaseFilename))
		if (err != nil) {
			color.Yellow("Error removing local temp database dump: %s", err)
		}
	},
}

func init() {
	syncCmd.AddCommand(filesCmd)
	syncCmd.AddCommand(dbCmd)
}


func validateArgs(args []string) {
	if (len(args) > 2) {
		panic("Too many arguments.  Command accepts only FROM and TO arguments")
	}
}