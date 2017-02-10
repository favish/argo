package project

import (
	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage individual argo projects.",
	Long: `Manipulate argo projects.`,
}

func init() {
	if blackfire := viper.GetBool("BLACKFIRE_SERVER_ID"); !blackfire {
		c := color.New(color.FgHiYellow).Add(color.Bold)
		c.Println("Warning: You do not have blackfire credentials stored in your environment! You will not be able to use blackfire until you add them!")
		c.Println("To add them easily, copy the export lines from https://blackfire.io/docs/integrations/docker (server are likely all you'll use) and add them to your ~/.zshrc \n")
	}
	ProjectCmd.AddCommand(createCmd)
	ProjectCmd.AddCommand(syncCmd)
	ProjectCmd.AddCommand(deleteCmd)
}