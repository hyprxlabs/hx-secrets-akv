/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/hyprxlabs/go/dotenv"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manages configuration for cli tool",
	Long: `Manages configuration for the hx-secrets-akv CLI tool.
	
Configuration can be used to set environment variables for the Azure SDK
such as AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_IDENTITY.  

AZURE_CLIENT_SECRET and AZURE_CLIENT_CERTIFICATE_PASSWORD will saved to the
operating system secret store if available. Otherwise they will not be saved.

Other configuration values will be stored as an env file in either 
$HOME/.config/hyprx/secrets/akv/.env or $XDG_CONFIG_HOME/hyprx/secrets/akv/.env

For windows this is stored in %APPDATA%/hyprx/secrets/akv/.env
	
	`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets a configuration value",
	Long:  `Sets a configuration value for the secrets store.`,
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		value := ""
		if len(args) != 2 {
			cmd.PrintErrf("Usage: %s set <name> <value>\n", cmd.Use)
			os.Exit(1)
		} else {
			name = args[0]
			value = args[1]
		}

		/*
			if name == "client.secret" {
				err := setOsSecret("CLIENT_SECRET", "hyprx-secrets-akv", value)
				if err != nil {
					cmd.PrintErrf("Error setting client secret: %v\n", err)
					os.Exit(CODE_ERROR)
				} else {
					value = "true"
				}
			}

			if name == "client.certificate.password" {
				err := setOsSecret("CERTIFICATE_PASSWORD", "hyprx-secrets-akv", value)
				if err != nil {
					cmd.PrintErrf("Error setting client certificate password: %v\n", err)
					os.Exit(CODE_ERROR)
				} else {
					value = "true"
				}
			}*/

		envName := ""

		switch name {
		case "tenant", "AZURE_TENANT_ID":
			envName = "AZURE_TENANT_ID"
		case "identity", "AZURE_IDENTITY":
			envName = "AZURE_IDENTITY"
		case "client.id", "AZURE_CLIENT_ID":
			envName = "AZURE_CLIENT_ID"
		case "client.secret", "AZURE_CLIENT_SECRET":
			envName = "AZURE_CLIENT_SECRET_KEY"
		case "client.certificate.password", "AZURE_CLIENT_CERTIFICATE_PASSWORD":
			envName = "AZURE_CLIENT_CERTIFICATE_PASSWORD_KEY"
		case "client.certificate.path", "AZURE_CLIENT_CERTIFICATE_PATH":
			envName = "AZURE_CLIENT_CERTIFICATE_PATH_KEY"
		}

		if envName == "" {
			cmd.PrintErr("Configuration key is not valid: " + name + "\n")
			os.Exit(CODE_ERROR)
		}

		dir := homeConfigDir()
		if dir == "" {
			cmd.PrintErr("Could not find home directory to store configuration.\n")
			os.Exit(CODE_ERROR)
		}

		envFile := filepath.Join(dir, ".env")
		exists := false
		var doc *dotenv.EnvDoc
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			doc = &dotenv.EnvDoc{}
		} else {
			exists = true
			bits, err := os.ReadFile(envFile)
			if err != nil {
				cmd.PrintErrf("Failed to read config file: %v\n", err)
				os.Exit(CODE_ERROR)
			}
			content := string(bits)
			doc, err = dotenv.Parse(content)
			if err != nil {
				cmd.PrintErrf("Failed to parse config file: %v\n", err)
				os.Exit(CODE_ERROR)
			}
		}

		doc.Set(envName, value)

		if exists {
			err := os.WriteFile(envFile, []byte(doc.String()), 0644)
			if err != nil {
				cmd.PrintErrf("Failed to save config: %v\n", err)
				os.Exit(CODE_ERROR)
			}
		} else {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0700); err != nil {
					cmd.PrintErrf("Failed to create config directory: %v\n", err)
					os.Exit(CODE_ERROR)
				}
			}

			err := os.WriteFile(envFile, []byte(doc.String()), 0644)
			if err != nil {
				cmd.PrintErrf("Failed to create config file: %v\n", err)
				os.Exit(CODE_ERROR)
			}
		}

		os.Exit(CODE_OK)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets a configuration value",
	Long:  `Gets a configuration value from the secrets store.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) != 1 {
			cmd.PrintErrf("Usage: %s get <name>\n", cmd.Use)
			os.Exit(1)
		} else {
			name = args[0]
		}

		envName := ""

		switch name {
		case "tenant", "AZURE_TENANT_ID":
			envName = "AZURE_TENANT_ID"
		case "identity", "AZURE_IDENTITY":
			envName = "AZURE_IDENTITY"
		case "client.id", "AZURE_CLIENT_ID":
			envName = "AZURE_CLIENT_ID"
		case "client.secret", "AZURE_CLIENT_SECRET":
			envName = "AZURE_CLIENT_SECRET_KEY"
		case "client.certificate.password", "AZURE_CLIENT_CERTIFICATE_PASSWORD":
			envName = "AZURE_CLIENT_CERTIFICATE_PASSWORD_KEY"
		case "client.certificate.path", "AZURE_CLIENT_CERTIFICATE_PATH":
			envName = "AZURE_CLIENT_CERTIFICATE_PATH_KEY"
		}

		if envName == "" {
			cmd.PrintErr("Configuration key is not valid: " + name + "\n")
			os.Exit(CODE_ERROR)
		}

		dir := homeConfigDir()
		if dir == "" {
			cmd.PrintErr("Could not find home directory to read configuration.\n")
			os.Exit(CODE_ERROR)
		}

		envFile := filepath.Join(dir, ".env")
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			cmd.PrintErrf("Configuration file does not exist: %s\n", envFile)
			os.Exit(CODE_ERROR)
		}

		bits, err := os.ReadFile(envFile)
		if err != nil {
			cmd.PrintErrf("Failed to read config file: %v\n", err)
			os.Exit(CODE_ERROR)
		}

		content := string(bits)
		doc, err := dotenv.Parse(content)
		if err != nil {
			cmd.PrintErrf("Failed to parse config file: %v\n", err)
			os.Exit(CODE_ERROR)
		}

		value, _ := doc.Get(envName)
		println(value)
		os.Exit(CODE_OK)
	},
}

var configRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Removes a configuration value",
	Long:  `Removes a configuration value from the secrets store.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) != 1 {
			cmd.PrintErrf("Usage: %s get <name>\n", cmd.Use)
			os.Exit(1)
		} else {
			name = args[0]
		}

		envName := ""

		switch name {
		case "tenant", "AZURE_TENANT_ID":
			envName = "AZURE_TENANT_ID"
		case "identity", "AZURE_IDENTITY":
			envName = "AZURE_IDENTITY"
		case "client.id", "AZURE_CLIENT_ID":
			envName = "AZURE_CLIENT_ID"
		case "client.secret", "AZURE_CLIENT_SECRET":
			envName = "AZURE_CLIENT_SECRET_KEY"
		case "client.certificate.password", "AZURE_CLIENT_CERTIFICATE_PASSWORD":
			envName = "AZURE_CLIENT_CERTIFICATE_PASSWORD_KEY"
		case "client.certificate.path", "AZURE_CLIENT_CERTIFICATE_PATH":
			envName = "AZURE_CLIENT_CERTIFICATE_PATH_KEY"
		}

		if envName == "" {
			cmd.PrintErr("Configuration key is not valid: " + name + "\n")
			os.Exit(CODE_ERROR)
		}

		dir := homeConfigDir()
		if dir == "" {
			cmd.PrintErr("Could not find home directory to read configuration.\n")
			os.Exit(CODE_ERROR)
		}

		envFile := filepath.Join(dir, ".env")
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			cmd.PrintErrf("Configuration file does not exist: %s\n", envFile)
			os.Exit(CODE_ERROR)
		}

		bits, err := os.ReadFile(envFile)
		if err != nil {
			cmd.PrintErrf("Failed to read config file: %v\n", err)
			os.Exit(CODE_ERROR)
		}

		content := string(bits)
		doc, err := dotenv.Parse(content)
		if err != nil {
			cmd.PrintErrf("Failed to parse config file: %v\n", err)
			os.Exit(CODE_ERROR)
		}

		doc2 := dotenv.NewDocument()
		for _, node := range doc.ToArray() {
			if node.Type == dotenv.VARIABLE_TOKEN && *node.Key != envName {
				doc2.Add(node)
			} else if node.Type != dotenv.VARIABLE_TOKEN {
				doc2.Add(node)
			}
		}

		content = doc2.String()
		err = os.WriteFile(envFile, []byte(content), 0644)
		if err != nil {
			cmd.PrintErrf("Failed to write config file: %v\n", err)
			os.Exit(CODE_ERROR)
		}

		os.Exit(CODE_OK)
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configRmCmd)
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
