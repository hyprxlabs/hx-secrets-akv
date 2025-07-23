/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	CODE_OK                        = 0
	CODE_ERROR                     = 1
	CODE_MISSING_VAULT_NAME        = 2
	CODE_INVALID_URL               = 3
	CODE_MISSING_VAULT_SECRET_NAME = 4
	CODE_INVALID_CREDENTIALS       = 10
	CODE_CLIENT_CREATION_FAILED    = 11
	CODE_SECRET_NOT_FOUND          = 12
	CODE_SECRET_EXPIRED            = 13
	CODE_SECRET_ROTATION_FAILED    = 14
	CODE_SECRET_GENERATE_FAILED    = 15
	CODE_SECRET_GET_FAILED         = 16
	CODE_SECRET_SET_FAILED         = 17
	CODE_SECRET_REMOVE_FAILED      = 18
	CODE_SECRET_LIST_FAILED        = 19
	CODE_SECRET_CONFIG_FAILED      = 20
	CODE_SECRET_CONFIG_NOT_FOUND   = 21
	CODE_OPERATION_CANCELLED       = 99
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hx-secrets-akv",
	Short: "A CLI tool for managing secrets in Azure Key Vault",
	Long: `A CLI tool for managing secrets in Azure Key Vault without the need to 
install the Azure CLI or use the Azure portal.
This tool allows you to list, add, remove, and manage secrets directly from the command line`,
	Version: "0.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.secrets-akv.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}
