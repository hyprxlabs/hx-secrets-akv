/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var completionsCmd = &cobra.Command{
	Use:       "completion [SHELL]",
	Short:     "Prints shell completion scripts",
	Long:      `Provides shell completion scripts for various shells like bash, zsh, fish, and powershell.`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Annotations: map[string]string{
		"commandType": "main",
	},
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			_ = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			_ = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
		}

		return nil
	},
}

func init() {
	completionsCmd.Hidden = true // Hide this command from the help output
	rootCmd.AddCommand(completionsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completionsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completionsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
