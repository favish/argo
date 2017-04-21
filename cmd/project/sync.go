package project

import (
	"github.com/spf13/cobra"
	"github.com/favish/argo/util"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"os"
	"fmt"
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
	Short: "Sync files FROM > TO.",
	Long: `Use this command to retrieve or push your files to and from remote environments.
	Ex. "argo project sync files dev local" brings dev files to your machine.
	Using rsync and project argo.yml configuration under the hood.
	Valid choices for FROM/TO are local, prod, or dev.

	WILL NOT SYNC TWO REMOTE HOSTS.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		validateArgs(args)

		destinationEnv := args[1]
		source, destination := getFileSyncPaths(args)

		if approve := util.GetApproval(fmt.Sprintf("This will sync files from %s (%s) to %s (%s), proceed?", args[0], source, args[1], destination)); !approve {
			os.Exit(1)
		} else {
			color.Cyan("Adding ssh aliases via gcloud for compute instances...")

			if (destination != "local") {
				color.Cyan(destinationEnv)
				destProject := projectConfig.GetString(fmt.Sprintf("environments.%s.project", destinationEnv))
				// TODO - only support local > remote or remote > local.
				util.ExecCmd("gcloud", "compute", "config-ssh", fmt.Sprintf("--project=%s", destProject))
			}

			util.ExecCmd("rsync", "-avzh", "--progress", source, destination)
		}
	},
}

var dbCmd = &cobra.Command{
	Use: "db",
	Short: "Sync database FROM > TO.",
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

		// Process:
		// Connect to SOURCE application pod
		// Dump database to temp file
		// Copy database to developer's machine
		// Unzip local copy
		// Connect to DESTINATION application pod
		// Create temp backup of DESTINATION database
		// Drop DESTINATION application database
		// Sync DESTINATION application database
		// Delete developer's temp file

		// Point kubectl at FROM, dump database to local temp file
		setKubectlConfig(source)
		// Get the container running application
		sourcePodName, err := util.ExecCmdChain("kubectl get pods --output='name' --selector='service=app' | awk -F'/' '{print $2}' | tr -d '\n'")
		if (err != nil) {
			color.Red("Error getting pod name: %s", err)
			os.Exit(1)
		}

		// Dump database on FROM environment
		color.Cyan("Creating a database dump on FROM container...")
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- /bin/bash -c 'cd docroot && drush sql-dump --gzip --result-file=/tmp/argo-db-tmp.sql'", sourcePodName))
		if (err != nil) {
			color.Red("Error getting dump: %s", err)
			os.Exit(1)
		}

		// Copy database to operator's machine
		color.Cyan("Copying dump from FROM container to your machine...")
		err = util.ExecCmd("kubectl", "cp", fmt.Sprintf("%s:/tmp/argo-db-tmp.sql.gz", sourcePodName), "/tmp/argo-db-tmp.sql.gz")
		if (err != nil) {
			color.Red("Error copying dump locally: %s", err)
			os.Exit(1)
		}

		// Unzip on operator's machine
		color.Cyan("Unzipping.  You may be prompted to overwrite, please do so.")
		err = util.ExecCmd("gunzip", "/tmp/argo-db-tmp.sql.gz")
		if (err != nil) {
			color.Red("Error unzipping database dump: %s", err)
			os.Exit(1)
		}

		// Point kubectl at DEST
		color.Cyan("Redirecting kubectl config to %s environment...", destination)
		setKubectlConfig(destination)

		// Get the container running application on destination
		destPodName, err := util.ExecCmdChain("kubectl get pods --output='name' --selector='service=app' | awk -F'/' '{print $2}' | tr -d '\n'")
		if (err != nil) {
			color.Red("Error getting pod name: %s", err)
			os.Exit(1)
		}

		color.Cyan("Creating backup of %s before dropping existing database, saving to /var/www/argo-db-tmp.sql.bak", destination)
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- /bin/bash -c 'cd docroot && drush sql-dump --gzip --result-file=/var/www/argo-db-tmp.sql.bak'", destPodName))
		if (err != nil) {
			color.Red("Error backing up: %s", err)
			os.Exit(1)
		}

		color.Cyan("Dropping %s database...", destination)
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s -- /bin/bash -c 'cd docroot && drush sql-drop -y'", destPodName))

		color.Cyan("Importing database dump to %s environment, may take a few moments...", destination)
		_, err = util.ExecCmdChain(fmt.Sprintf("kubectl exec --stdin=true --tty=true %s < /tmp/argo-db-tmp.sql -- /bin/bash -c 'cd docroot && drush sqlc'", destPodName))
		if (err != nil) {
			color.Red("Error importing database: %s", err)
			os.Exit(1)
		}

		err = util.ExecCmd("rm", "/tmp/argo-db-tmp.sql")
		err = util.ExecCmd("rm", "/tmp/argo-db-tmp.sql.gz")
		if (err != nil) {
			color.Red("Error removing temp database: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	syncCmd.AddCommand(filesCmd)
	syncCmd.AddCommand(dbCmd)
}


func validateArgs(args []string) {
	if (len(args) > 2) {
		color.Red("Too many args.  Command accepts only FROM and TO arguments")
	}

	validArgs := map[string]bool {
		"dev": true,
		"local": true,
		"prod": true,
	}
	// Validate arguments
	for _, arg := range args {
		if !validArgs[arg] {
			color.Red("Invalid argument provided.  Must be one of: local, prod, dev.")
			os.Exit(1)
		}
	}
}

func getFileSyncPaths(args []string) (string, string) {
	projectName := projectConfig.GetString("project-name")
	locations := map[string]string {
		"dev": fmt.Sprintf("%s:%s%s/", viper.GetString("environments.dev.files-instance"), "/mnt/disks/", projectName),
		"prod": fmt.Sprintf("%s:%s%s/", viper.GetString("environments.prod.files-instance"), "/mnt/disks/", projectName),
		"local": "docroot/sites/default/files/",
	}

	from := locations[args[0]]
	to := locations[args[1]]

	return from, to
}