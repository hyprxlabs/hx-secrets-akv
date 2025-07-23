/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/spf13/cobra"
)

type Secret struct {
	Key         string             `json:"key"`
	Value       string             `json:"value"`
	Tags        map[string]*string `json:"tags,omitempty"`
	Enabled     bool               `json:"enabled,omitempty"`
	Version     string             `json:"version,omitempty"`
	ExpiresAt   string             `json:"expires_at,omitempty"`
	StartsAt    string             `json:"starts_at,omitempty"`
	ContentType string             `json:"content_type,omitempty"`
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets the keyvault record from the secrets store as JSON",
	Long: `Gets a secret value from azure key vault and prints it to stdout in JSON format.
	
The URL argument is optional. If provided, it should be in the format:
https://<vault-name>.vault.azure.net/secrets/<key-name>/[<version>]
akv://<vault-name>/<key-name>[/<version>]

If the URL is not provided, you must specify the vault and key using flags.`,
	Run: func(cmd *cobra.Command, args []string) {

		vaultName, _ := cmd.Flags().GetString("vault")
		key, _ := cmd.Flags().GetString("key")
		version, _ := cmd.Flags().GetString("version")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		logDebug, _ = cmd.Flags().GetBool("debug")
		ur := ""
		if len(args) > 0 {
			ur = args[0]
		}

		if len(ur) > 0 {
			uri, err := url.Parse(ur)
			if uri.Scheme != "https" && uri.Scheme != "akv" {
				if logDebug {
					cmd.PrintErrf("Invalid URL scheme: %s\n", uri.Scheme)
				}
				os.Exit(CODE_INVALID_URL)
			}

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
					if len(parts) > 1 {
						version = parts[1]
					}
				} else {
					key = path
				}
			}
		}

		if vaultName == "" {
			if !logDebug {
				cmd.PrintErrf("Vault name is required. Use --vault <name> option or the URL argument.\n")
			}
			os.Exit(CODE_MISSING_VAULT_NAME)
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

		resp, err := client.GetSecret(context.TODO(), key, version, nil)
		var respErr *azcore.ResponseError
		if err != nil || errors.As(err, &respErr) {
			if respErr.ErrorCode == "SecretNotFound" {
				if !logDebug {
					cmd.PrintErrf("Secret not found: %s\n", key)
				}
				os.Exit(CODE_SECRET_NOT_FOUND)
			} else {
				if logDebug {
					cmd.PrintErrf("Failed to get secret: %v\n", err)
				}
				os.Exit(CODE_SECRET_GET_FAILED)
			}
		}

		if err != nil {
			if logDebug {
				cmd.PrintErrf("Failed to get secret: %v\n", err)
			}
			os.Exit(CODE_SECRET_NOT_FOUND)
		}

		expires := ""
		if resp.Attributes.Expires != nil {
			expires = resp.Attributes.Expires.Format("2006-01-02T15:04:05Z07:00")
		}
		startsAt := ""
		if resp.Attributes.NotBefore != nil {
			startsAt = resp.Attributes.NotBefore.Format("2006-01-02T15:04:05Z07:00")
		}

		contentType := ""
		if resp.ContentType != nil {
			contentType = *resp.ContentType
		}

		secret := Secret{
			Key:         resp.ID.Name(),
			Value:       *resp.Value,
			ContentType: contentType,
			Tags:        resp.Tags,
			Enabled:     *resp.Attributes.Enabled,
			Version:     resp.ID.Version(),
			ExpiresAt:   expires,
			StartsAt:    startsAt,
		}

		bytes, err := json.Marshal(secret)
		if err != nil {
			if logDebug {
				cmd.PrintErrf("Failed to marshal secret: %v\n", err)
			}
			os.Exit(CODE_SECRET_GET_FAILED)
		}

		cmd.OutOrStdout().Write(bytes)
		cmd.OutOrStdout().Write([]byte("\n"))
	},
}

var getValueCmd = &cobra.Command{
	Use:   "value [URL]",
	Short: "Get a value from azure key vault",
	Long: `Get a value from an azure key vault.

The URL argument is optional. If provided, it should be in the format:
https://<vault-name>.vault.azure.net/secrets/<key-name>/[<version>]
akv://<vault-name>/<key-name>[/<version>]

If the URL is not provided, you must specify the vault and key using flags.
	`,
	Example: `hx-secrets-akv get value --vault myvault --key mykey
hx-secrets-akv get value https://myvault.vault.azure.net/secrets/mykey/1234567890abcdef
hx-secrets-akv get value akv://myvault/mykey	
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		vaultName, _ := cmd.Flags().GetString("vault")
		key, _ := cmd.Flags().GetString("key")
		version, _ := cmd.Flags().GetString("version")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		logDebug, _ := cmd.Flags().GetBool("debug")
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
					if len(parts) > 1 {
						version = parts[1]
					}
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

		resp, err := client.GetSecret(context.TODO(), key, version, nil)
		var respErr *azcore.ResponseError
		if err != nil || errors.As(err, &respErr) {
			if respErr.ErrorCode == "SecretNotFound" {
				if !logDebug {
					cmd.PrintErrf("Secret not found: %s\n", key)
				}
				os.Exit(CODE_SECRET_NOT_FOUND)
			} else {
				if logDebug {
					cmd.PrintErrf("Failed to get secret: %v\n", err)
				}
				os.Exit(CODE_SECRET_GET_FAILED)
			}
		}

		if err != nil {
			if logDebug {
				cmd.PrintErrf("Failed to get secret: %v\n", err)
			}
			os.Exit(CODE_SECRET_GET_FAILED)
		}
		cmd.Println(*resp.Value)
		os.Exit(0)
	},
}

func init() {
	getValueCmd.Flags().StringP("vault", "v", "", "Key Vault name (e.g., myvault)")
	getValueCmd.Flags().StringP("key", "k", "", "Key name in the Key Vault")
	getValueCmd.Flags().StringP("version", "V", "", "Version of the key (optional)")
	getValueCmd.Flags().BoolP("interactive", "i", false, "Use interactive authentication")
	getValueCmd.Flags().Bool("device-code", false, "Use device code authentication")
	getValueCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages")

	getCmd.Flags().StringP("vault", "v", "", "Key Vault name (e.g., myvault)")
	getCmd.Flags().StringP("key", "k", "", "Key name in the Key Vault")
	getCmd.Flags().StringP("version", "V", "", "Version of the key (optional)")
	getCmd.Flags().BoolP("interactive", "i", false, "Use interactive authentication")
	getCmd.Flags().Bool("device-code", false, "Use device code authentication")
	getCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages")
	getCmd.AddCommand(getValueCmd)

	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
