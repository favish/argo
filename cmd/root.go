package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/favish/argo/cmd/components"
	"github.com/favish/argo/cmd/project"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "argo",
	Short: "Argo allows developers to quickly get started developing and configuring Favish projects.",
	Long: `
		Use 'argo components' to install/uninstall sub components and 'argo project' to manipulate projects.
	`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {

	cobra.OnInitialize(initConfig)

	RootCmd.AddCommand(versionCmd)

	RootCmd.AddCommand(components.ComponentsCmd)
	RootCmd.AddCommand(project.ProjectCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().Bool("debug", false, "Run in debug mode.  Increases stdout verbosity.")
	RootCmd.PersistentFlags().BoolP("auto-yes", "y", false, "Answer yes to all prompts.")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("auto-yes", RootCmd.PersistentFlags().Lookup("auto-yes"))

	//RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.favish-cloud.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
// TODO - Global argo configuration - MEA
// TODO - Configuration for projects now set in ProjectCmd to a scoped viper instance.  This master instance no longer used
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("argo") // name of config file (without extension)
	viper.AddConfigPath(".")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// TODO - Set which config file is used if it becomes necessary and only log it when needed (argo project commands) - MEA
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		viper.Set("noConfig", true)
	}
}
