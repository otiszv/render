package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var (
	version   string
	buildDate string
)

var RootCmd = &cobra.Command{
	Use:   "jenkinsfilext",
	Short: "Welcome to jenkinsfilext",
	Long:  `jenkinsfilext is Devops's jenkinsfile render engine.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of jenkinsfilext",
	Long:  `Print the version number of jenkinsfilext`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("JenkinsfileXT Version: %s(%s)\n", version, buildDate)
	},
}

func Execute(_version, _buildDate string) {
	version = _version
	buildDate = _buildDate

	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.AddCommand(versionCmd)
}

func initConfig() {
}
