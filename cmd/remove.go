/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "rm",
	Aliases: []string{"remove"},
	Short:   "Removes a secret from the Azure Key Vault",
	Long: `Removes a secret from the Azure Key Vault.

	If --puge is specified, the secret will be permanently deleted.
	If --force is specified, the command will not prompt for confirmation before deleting the secret.`,

	Run: func(cmd *cobra.Command, args []string) {
		vaultName, _ := cmd.Flags().GetString("vault")
		key, _ := cmd.Flags().GetString("key")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		purge, _ := cmd.Flags().GetBool("purge")
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
					parts := strings.SplitN(path, "/", 2)
					key = parts[0]
				} else {
					key = path
				}
			}
		}

		if vaultName == "" {
			if logDebug {
				cmd.PrintErrf("Vault name is required\n")
			}
			os.Exit(CODE_MISSING_VAULT_NAME)
		}

		if key == "" {
			if logDebug {
				cmd.PrintErrf("Key name is required\n")
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
			if logDebug {
				cmd.PrintErrf("Failed to create client: %v\n", err)
			}
			os.Exit(CODE_CLIENT_CREATION_FAILED)
		}

		if !force {
			fmt.Println("Delete secret [y/n]:")
			confirm := ""
			for confirm != "y" && confirm != "n" {
				fmt.Scanln(&confirm)
				if confirm == "n" {
					if logDebug {
						cmd.Println("Operation cancelled.")
					}
					os.Exit(CODE_OPERATION_CANCELLED)
				} else if confirm != "y" {
					cmd.PrintErrf("Invalid input. Please enter 'y' or 'n'.\n")
				}
			}
		}

		resp, err := client.DeleteSecret(cmd.Context(), key, nil)
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == "SecretNotFound" {
			if logDebug {
				cmd.Printf("Secret %s already removed from %s.\n", key, vaultName)
			}
			os.Exit(CODE_OK)
		}

		if err != nil {
			if logDebug {
				cmd.PrintErrf("Failed to delete secret: %v\n", err)
			}
			os.Exit(CODE_SECRET_REMOVE_FAILED)
		}

		if resp.ScheduledPurgeDate != nil && purge {
			pager := client.NewListDeletedSecretPropertiesPager(nil)
			for pager.More() {
				page, err := pager.NextPage(context.TODO())
				if err != nil {
					// TODO: handle error
				}
				for _, secret := range page.Value {
					if secret.ID.Name() != key {
						continue
					}

					_, err := client.PurgeDeletedSecret(context.TODO(), secret.ID.Name(), nil)
					if err != nil {
						if !logDebug {
							cmd.PrintErrf("Failed to purge secret: %v\n", err)
						}
						os.Exit(CODE_SECRET_REMOVE_FAILED)
					}
				}
			}
		}

		os.Exit(CODE_OK)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Define flags for the remove command
	removeCmd.Flags().StringP("vault", "v", "", "The name of the Azure Key Vault")
	removeCmd.Flags().StringP("key", "k", "", "The name of the secret to remove")
	removeCmd.Flags().StringP("version", "V", "", "The version of the secret to remove")
	removeCmd.Flags().BoolP("interactive", "i", false, "Use interactive browser login")
	removeCmd.Flags().BoolP("device-code", "D", false, "Use device code authentication")
	removeCmd.Flags().BoolP("force", "f", false, "Force removal without confirmation")
	removeCmd.Flags().BoolP("purge", "p", false, "Permanently delete the secret")
	removeCmd.Flags().BoolP("debug", "d", false, "Enable debug output")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
