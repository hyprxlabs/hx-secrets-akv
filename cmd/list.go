/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Example: `hx-secrets-akv ls --vault myvault
hx-secrets-akv ls --vault myvault --query mysecret*
hx-secrets-akv ls https://myvault.vault.azure.net/secrets/mysecret*
hx-secrets-akv ls akv://myvault/mysecret*`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultName, _ := cmd.Flags().GetString("vault")
		query, _ := cmd.Flags().GetString("query")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		logDebug, _ = cmd.Flags().GetBool("debug")
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
				query = path
			}
		}

		if len(vaultName) == 0 {
			if !logDebug {
				cmd.PrintErrf("vault name is required. Use --vault <name> option or the URL argument.\n")
			}
			os.Exit(5)
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

		resp := client.NewListSecretPropertiesPager(nil)

		for resp.More() {
			page, err := resp.NextPage(cmd.Context())
			if err != nil {
				if logDebug {
					cmd.PrintErrf("Failed to get secret: %v\n", err)
				}
				os.Exit(CODE_SECRET_LIST_FAILED)
			}
			for _, secret := range page.Value {
				name := secret.ID.Name()
				match := len(query) == 0

				if !match {
					res, _ := filepath.Match(query, name)
					match = res
				}

				if !match {
					continue
				}

				cmd.Println(name)
			}
		}

		os.Exit(CODE_OK)
	},
}

func init() {
	listCmd.Flags().StringP("vault", "v", "", "The name of the Azure Key Vault")
	listCmd.Flags().StringP("query", "s", "", "A query to filter the secrets by name")
	listCmd.Flags().BoolP("interactive", "i", false, "Use interactive authentication")
	listCmd.Flags().Bool("device-code", false, "Use device code authentication")
	listCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages")

	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
