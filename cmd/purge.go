/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/spf13/cobra"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultName, _ := cmd.Flags().GetString("vault")
		key, _ := cmd.Flags().GetString("key")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		logDebug, _ = cmd.Flags().GetBool("debug")
		force, _ := cmd.Flags().GetBool("force")

		ur := ""
		if len(args) > 0 {
			ur = args[0]
		}

		if len(ur) > 0 {
			uri, err := url.Parse(ur)
			if err != nil {
				if logDebug {
					cmd.PrintErrf("Invalid URL: %s\n", ur)
				}
				os.Exit(CODE_INVALID_URL)
			}

			vaultName = uri.Host
			path := uri.Path
			path = strings.TrimPrefix(path, "/")
			path = strings.TrimSuffix(path, "/")
			path = strings.TrimPrefix(path, "secrets/")
			if len(path) > 0 {
				if strings.ContainsRune(path, '/') {
					parts := strings.Split(path, "/")
					key = parts[0]
				} else {
					key = path
				}
			}
		}

		if vaultName == "" {
			if !logDebug {
				cmd.PrintErrf("Vault name is required.\n")
			}
			os.Exit(CODE_MISSING_VAULT_NAME)
		}

		if key == "" {
			if !logDebug {
				cmd.PrintErrf("Key name is required.\n")
			}
			os.Exit(CODE_MISSING_VAULT_SECRET_NAME)
		}

		if !strings.HasSuffix(vaultName, ".vault.azure.net") {
			vaultName += ".vault.azure.net"
		}

		inter := ""
		if deviceCode {
			inter = "device-code"
		}

		if interactive {
			inter = "interactive"
		}

		var interactivePtr *string
		if inter != "" {
			interactivePtr = &inter
		}

		creds, err := getCredential(interactivePtr, cmd.Context())
		if err != nil {
			if logDebug {
				cmd.PrintErrf("Failed to get credentials: %v\n", err)
			}
			os.Exit(CODE_INVALID_CREDENTIALS)
		}

		client, err := azsecrets.NewClient("https://"+vaultName, creds, nil)
		if err != nil {
			if !logDebug {
				cmd.PrintErrf("Failed to create client: %v\n", err)
			}
			os.Exit(CODE_CLIENT_CREATION_FAILED)
		}

		if !force {
			fmt.Println("Purge secret [y/n]:")
			confirm := ""
			for confirm != "y" && confirm != "n" {
				fmt.Scanln(&confirm)
				if confirm == "n" {
					if !logDebug {
						cmd.Println("Operation cancelled.")
					}
					os.Exit(CODE_OPERATION_CANCELLED)
				} else if confirm != "y" {
					cmd.PrintErrf("Invalid input. Please enter 'y' or 'n'.\n")
				}
			}
		}

		_, err = client.PurgeDeletedSecret(cmd.Context(), key, nil)
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == "SecretNotFound" {
			if logDebug {
				cmd.Printf("Secret %s already purged %s.\n", key, vaultName)
			}
			os.Exit(CODE_OK)
		}

		if err != nil {
			if logDebug {
				cmd.PrintErrf("Failed to purge secret: %v\n", err)
			}
			os.Exit(CODE_SECRET_REMOVE_FAILED)
		}

		os.Exit(CODE_OK)
	},
}

func init() {
	purgeCmd.Flags().StringP("vault", "v", "", "Name of the Azure Key Vault")
	purgeCmd.Flags().StringP("key", "k", "", "Key of the secret to purge")
	purgeCmd.Flags().BoolP("interactive", "i", false, "Use interactive authentication")
	purgeCmd.Flags().BoolP("device-code", "D", false, "Use device code authentication")
	purgeCmd.Flags().BoolP("debug", "d", false, "Enable debug output")
	purgeCmd.Flags().BoolP("force", "f", false, "Force purge without confirmation")
	// Note: The purge flag is not used in this command,

	rootCmd.AddCommand(purgeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// purgeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// purgeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
