/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/hyprxlabs/go/env"
	"github.com/mashiike/longduration"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets azure key vault secret parameters",
	Long: `
Sets the parameters for a secret in Azure Key Vault.

This command allows you to set a secret value, expiration time, start time, and tags.
You can also read the secret value from a file, environment variable, or stdin.

You can specify the vault name, key, and various parameters using flags or by providing a URL.
If the URL is not provided, you must specify the vault and key using flags.
	`,
	Example: `hx-secrets-akv set --vault myvault --key mykey --value myvalue
hx-secrets-akv set https://myvault.vault.azure.net/secrets/mykey --value-file myvalue.txt
hx-secrets-akv set akv://myvault/mykey --value-variable MY_SECRET_VAR
echo "myvalue" | hx-secrets-akv set --vault myvault --key mykey --stdin
hx-secrets-akv set --vault myvault --key mykey --expires-at 2025-12-31T23:59:59Z --not-before 2025-01`,
	Run: func(cmd *cobra.Command, args []string) {

		vaultName, _ := cmd.Flags().GetString("vault")
		key, _ := cmd.Flags().GetString("key")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		quiet, _ := cmd.Flags().GetBool("quiet")

		value, _ := cmd.Flags().GetString("value")
		valueVar, _ := cmd.Flags().GetString("value-variable")
		valueFile, _ := cmd.Flags().GetString("value-file")
		valueStdin, _ := cmd.Flags().GetBool("stdin")
		expiresAt, _ := cmd.Flags().GetString("expires-at")
		startsAt, _ := cmd.Flags().GetString("starts-at")
		tags, _ := cmd.Flags().GetString("tag")

		var valuePtr *string
		if len(value) > 0 {
			valuePtr = &value
		}

		if valueStdin {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				if !quiet {
					cmd.PrintErrf("Error reading from stdin: %v\n", err)
				}
				os.Exit(1)
			}
			value := string(bytes)
			value = strings.TrimSpace(value)
			valuePtr = &value
		}

		if len(valueFile) > 0 {
			bytes, err := os.ReadFile(valueFile)
			if err != nil {
				if !quiet {
					cmd.PrintErrf("Error reading file %s: %v\n", valueFile, err)
				}
				os.Exit(1)
			}
			value := string(bytes)
			if len(value) > 0 {
				valuePtr = &value
			}
		}

		if len(valueVar) > 0 {
			value := env.Get(valueVar)
			if value != "" {
				valuePtr = &value
			}
		}

		ur := ""
		if len(args) > 0 {
			ur = args[0]
		}

		if len(ur) > 0 {
			uri, err := url.Parse(ur)
			if err != nil {
				if !quiet {
					cmd.PrintErrf("Invalid URL: %s\n", ur)
				}
				os.Exit(4)
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
			if !quiet {
				cmd.PrintErrf("Failed to get credentials: %v\n", err)
			}
			os.Exit(1)
		}

		client, err := azsecrets.NewClient("https://"+vaultName, creds, nil)
		if err != nil {
			if !quiet {
				cmd.PrintErrf("Failed to create client: %v\n", err)
			}
			os.Exit(2)
		}

		params := &azsecrets.SetSecretParameters{}
		if valuePtr != nil {
			params.Value = valuePtr
		}
		if len(expiresAt) > 0 {
			dur, err := longduration.ParseDuration(expiresAt)
			if err == nil {

				if params.SecretAttributes == nil {
					params.SecretAttributes = &azsecrets.SecretAttributes{}
				}
				dt := time.Now().Add(dur)
				params.SecretAttributes.Expires = &dt
			} else {
				targetTime, err := time.Parse(time.RFC3339, expiresAt)
				if err == nil {
					if params.SecretAttributes == nil {
						params.SecretAttributes = &azsecrets.SecretAttributes{}
					}
					params.SecretAttributes.Expires = &targetTime
				}
			}
		}

		if len(startsAt) > 0 {
			dur, err := longduration.ParseDuration(startsAt)
			if err == nil {
				if params.SecretAttributes == nil {
					params.SecretAttributes = &azsecrets.SecretAttributes{}
				}
				dt := time.Now().Add(dur)
				params.SecretAttributes.NotBefore = &dt
			} else {
				targetTime, err := time.Parse(time.RFC3339, startsAt)
				if err == nil {
					if params.SecretAttributes == nil {
						params.SecretAttributes = &azsecrets.SecretAttributes{}
					}
					params.SecretAttributes.NotBefore = &targetTime
				}
			}
		}

		if len(tags) > 0 {
			params.Tags = make(map[string]*string)
			for _, tag := range strings.Split(tags, ",") {
				parts := strings.SplitN(tag, "=", 2)
				if len(parts) == 2 {
					params.Tags[parts[0]] = &parts[1]
				} else {
					params.Tags[parts[0]] = nil
				}
			}
		}

		resp, err := client.SetSecret(cmd.Context(), key, *params, nil)
		if err != nil {
			if !quiet {
				cmd.PrintErrf("Failed to set secret: %v\n", err)
			}
			os.Exit(1)
		}

		if !quiet {
			cmd.Println("Secret set successfully. version: " + resp.ID.Version())
		}

		os.Exit(0)
	},
}

var setValueCmd = &cobra.Command{
	Use:   "value",
	Short: "Set the value of a secret in Azure Key Vault",
	Long: `Set the value of a secret in Azure Key Vault.
	
Only the value of the secret is set. Other parameters like expiration time, start time,
and tags can be set using the 'set' command.

You can specify the vault name, and key using flags or by providing a URL.
If the URL is not provided, you must specify the vault and key using flags.
	`,
	Example: `hx-secrets-akv set value --vault myvault --key mykey --value myvalue
hx-secrets-akv set value https://myvault.vault.azure.net/secrets/mykey --value-file myvalue.txt
hx-secrets-akv set value akv://myvault/mykey --value-variable MY_SECRET_VAR
echo "myvalue" | hx-secrets-akv set value --vault myvault --key mykey --stdin
	`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultName, _ := cmd.Flags().GetString("vault")
		key, _ := cmd.Flags().GetString("key")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		logDebug, _ = cmd.Flags().GetBool("debug")

		value, _ := cmd.Flags().GetString("value")
		valueVar, _ := cmd.Flags().GetString("value-variable")
		valueFile, _ := cmd.Flags().GetString("value-file")
		valueStdin, _ := cmd.Flags().GetBool("stdin")
		var valuePtr *string
		if len(args) > 1 {
			valuePtr = &args[1]
		}

		if len(value) > 0 {
			if valuePtr != nil {
				valuePtr = &value
			}
		}

		if valuePtr == nil && valueStdin {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				if logDebug {
					cmd.PrintErrf("Error reading from stdin: %v\n", err)
				}
				os.Exit(CODE_ERROR)
			}
			value := string(bytes)
			value = strings.TrimSpace(value)
			valuePtr = &value
		}

		if valuePtr == nil && len(valueFile) > 0 {
			bytes, err := os.ReadFile(valueFile)
			if err != nil {
				if logDebug {
					cmd.PrintErrf("Error reading file %s: %v\n", valueFile, err)
				}
				os.Exit(CODE_ERROR)
			}
			value := string(bytes)
			if len(value) > 0 {
				valuePtr = &value
			}
		}

		if valuePtr == nil && len(valueVar) > 0 {
			value := env.Get(valueVar)
			if value != "" {
				valuePtr = &value
			}
		}

		if valuePtr == nil {
			if !logDebug {
				cmd.PrintErrf("Value must be specified. Use --value, --value-variable, --value-file, or provide it as an argument.\n")
			}
			os.Exit(CODE_ERROR)
		}

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
				cmd.PrintErrf("Vault name must be specified.\n")
			}
			os.Exit(CODE_MISSING_VAULT_NAME)
		}

		if key == "" {
			if logDebug {
				cmd.PrintErrf("Key name must be specified.\n")
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

		params := &azsecrets.SetSecretParameters{}
		if valuePtr != nil {
			params.Value = valuePtr
		}

		resp, err := client.SetSecret(cmd.Context(), key, *params, nil)
		if err != nil {
			if !logDebug {
				cmd.PrintErrf("Failed to set secret: %v\n", err)
			}
			os.Exit(CODE_SECRET_SET_FAILED)
		}

		if !logDebug {
			cmd.Println("Secret set successfully. version: " + resp.ID.Version())
		}

		os.Exit(0)
	},
}

func init() {
	setCmd.Flags().StringP("vault", "v", "", "Azure Key Vault name (without .vault.azure.net)")
	setCmd.Flags().StringP("key", "k", "", "Key name in the Key Vault")
	setCmd.Flags().BoolP("interactive", "i", false, "Use interactive login")
	setCmd.Flags().BoolP("device-code", "D", false, "Use device code authentication")
	setCmd.Flags().BoolP("debug", "d", false, "Suppress output messages")
	setCmd.Flags().StringP("value", "V", "", "Value of the secret")
	setCmd.Flags().StringP("value-variable", "a", "", "Read value from an environment variable")
	setCmd.Flags().StringP("value-file", "f", "", "Read value from a file")
	setCmd.Flags().BoolP("stdin", "s", false, "Read value from stdin")
	setCmd.Flags().StringP("expires-at", "e", "", "Expiration time of the secret (RFC3339 or duration format)")
	setCmd.Flags().StringP("not-before", "b", "", "Start time of the secret (RFC3339 or duration format)")
	setCmd.Flags().StringArrayP("tag", "t", nil, "Tags for the secret in key=value format. Multiple tags can be specified with multiple -t flags.")

	setValueCmd.Flags().StringP("vault", "v", "", "Azure Key Vault name (without .vault.azure.net)")
	setValueCmd.Flags().StringP("key", "k", "", "Key name in the Key Vault")
	setValueCmd.Flags().BoolP("interactive", "i", false, "Use interactive login")
	setValueCmd.Flags().BoolP("device-code", "D", false, "Use device code authentication")
	setValueCmd.Flags().BoolP("debug", "d", false, "Suppress output messages")
	setValueCmd.Flags().StringP("value", "V", "", "Value of the secret")
	setValueCmd.Flags().StringP("value-variable", "a", "", "Environment variable containing the value")
	setValueCmd.Flags().StringP("value-file", "f", "", "File containing the value")
	setValueCmd.Flags().BoolP("stdin", "s", false, "Read value from stdin")

	setCmd.AddCommand(setValueCmd)

	rootCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
