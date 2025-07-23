/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/hyprxlabs/go/secrets"
	"github.com/spf13/cobra"
)

var logDebug = false

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Gets or sets the keyvault record from the secrets store",
	Long: `Gets a secret value from azure key vault and prints it to stdout. 
	If the secret does not exist, it will create a new generated secret with the given key.
	This command is useful for retrieving secrets in a secure manner without exposing them in the command line.`,
	Run: func(cmd *cobra.Command, args []string) {

		vaultName, _ := cmd.Flags().GetString("vault")
		key, _ := cmd.Flags().GetString("key")
		version, _ := cmd.Flags().GetString("version")
		interactive, _ := cmd.Flags().GetBool("interactive")
		deviceCode, _ := cmd.Flags().GetBool("device-code")
		logDebug, _ = cmd.Flags().GetBool("debug")
		upper, _ := cmd.Flags().GetBool("upper")
		lower, _ := cmd.Flags().GetBool("lower")
		digits, _ := cmd.Flags().GetBool("digits")
		noSpecial, _ := cmd.Flags().GetBool("no-special")
		nist, _ := cmd.Flags().GetBool("nist")
		special, _ := cmd.Flags().GetString("special")
		chars, _ := cmd.Flags().GetString("chars")
		size, _ := cmd.Flags().GetInt16("size")
		ur := ""
		if len(args) > 0 {
			ur = args[0]
		}

		if nist {
			upper = true
			lower = true
			digits = true
			noSpecial = false
			special = "@#`~_-[]|+="
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
			if !logDebug {
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
		rotate := false
		resp, err := client.GetSecret(context.TODO(), key, version, nil)
		if err == nil {
			if resp.Attributes.Expires != nil && !time.Now().Before(*resp.Attributes.Expires) {

				truish := "true"
				if resp.Tags != nil && resp.Tags["auto-rotate"] != nil && *resp.Tags["auto-rotate"] == truish {
					rotate = true
				}

				if !rotate {
					if logDebug {
						cmd.PrintErrf("Secret %s has expired and is not tagged for rotation.\n", key)
					}
					os.Exit(CODE_SECRET_EXPIRED)
				}

			} else {
				println(*resp.Value)
				os.Exit(0)
			}
		}

		// cast to azcore.ResponseError to check if the secret does not exist
		if err != nil || rotate {
			var respErr *azcore.ResponseError
			if rotate || (errors.As(err, &respErr) && respErr.ErrorCode == "SecretNotFound") {
				if logDebug {
					cmd.Println("Secret not found, creating a new secret with a generated value.")
				} // Create a new secret with a generated value

				generatedValue := ""
				if len(chars) > 0 {
					generatedValue, err = secrets.Generate(size, secrets.WithChars(chars), secrets.WithValidator(func(s []rune) error {
						return nil
					}))
					if err != nil {
						if !logDebug {
							cmd.PrintErrf("Failed to generate secret: %v\n", err)
						}
						os.Exit(CODE_SECRET_ROTATION_FAILED)
					}
				} else {

					var symbolOpt secrets.SetOption
					if !noSpecial {
						symbolOpt = secrets.WithSymbols(special)
					} else {
						symbolOpt = secrets.WithNoSymbols()
					}

					validator := secrets.WithValidator(func(s []rune) error {
						requireUpper := upper
						requireLower := lower
						requireDigits := digits
						requireSpecial := !noSpecial
						hasUpper := false
						hasLower := false
						hasDigits := false
						hasSpecial := false

						for _, r := range s {
							if unicode.IsUpper(r) {
								hasUpper = true
							} else if unicode.IsLower(r) {
								hasLower = true
							} else if unicode.IsDigit(r) {
								hasDigits = true
							} else {
								hasSpecial = true
							}
						}

						if requireUpper && !hasUpper {
							return errors.New("secret must contain at least one uppercase letter")
						}

						if requireLower && !hasLower {
							return errors.New("secret must contain at least one lowercase letter")
						}

						if requireDigits && !hasDigits {
							return errors.New("secret must contain at least one digit")
						}

						if requireSpecial && !hasSpecial {
							return errors.New("secret must contain at least one special character")
						}

						if len(s) < 1 {
							return errors.New("secret must be at least 1 character long")
						}

						return nil
					})

					generatedValue, err = secrets.Generate(size, symbolOpt, validator, secrets.WithUpper(upper), secrets.WithLower(lower), secrets.WithDigits(digits))
				}

				if err != nil {
					if logDebug {
						cmd.PrintErrf("Failed to generate secret: %v\n", err)
					}
					os.Exit(CODE_SECRET_GENERATE_FAILED)
				}

				params := &azsecrets.SetSecretParameters{}
				params.Value = &generatedValue
				_, err = client.SetSecret(cmd.Context(), key, *params, nil)
				if err != nil {
					if logDebug {
						cmd.PrintErrf("Failed to set secret: %v\n", err)
					}
					os.Exit(CODE_SECRET_SET_FAILED)
				}

				println(generatedValue)
				os.Exit(0)
			}
		}
	},
}

func init() {
	resolveCmd.Flags().StringP("vault", "v", "", "Azure Key Vault name (e.g., myvault.vault.azure.net)")
	resolveCmd.Flags().StringP("key", "k", "", "The key of the secret to resolve")
	resolveCmd.Flags().BoolP("interactive", "i", false, "Use interactive authentication")
	resolveCmd.Flags().BoolP("device-code", "D", false, "Use device code authentication")
	resolveCmd.Flags().BoolP("debug", "d", false, "Suppress output messages")
	resolveCmd.Flags().BoolP("upper", "u", false, "Require at least one uppercase letter")
	resolveCmd.Flags().BoolP("lower", "l", false, "Require at least one lowercase letter")
	resolveCmd.Flags().BoolP("digits", "g", false, "Require at least one digit")
	resolveCmd.Flags().BoolP("no-special", "n", false, "Do not require special characters")
	resolveCmd.Flags().BoolP("nist", "N", false, "Use NIST compliant password generation (upper, lower, digits, special characters)")
	resolveCmd.Flags().StringP("special", "s", "@#`~_-[]|+=", "Special characters to use in the secret")
	resolveCmd.Flags().StringP("chars", "c", "", "Custom characters to use in the secret")
	resolveCmd.Flags().Int16P("size", "z", 16, "Size of the generated secret (default is 32 characters)")

	rootCmd.AddCommand(resolveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resolveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resolveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
