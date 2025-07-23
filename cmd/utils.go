package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hyprxlabs/go/dotenv"
	"github.com/hyprxlabs/go/env"
)

func getCredential(interactive *string, ctx context.Context) (azcore.TokenCredential, error) {

	targetDir := homeConfigDir()
	if targetDir == "" {
		targetDir = osConfigDir()
	}

	if targetDir != "" {
		target := filepath.Join(targetDir, ".env")
		if _, err := os.Stat(target); err == nil {
			bits, err := os.ReadFile(target)
			if err == nil {
				content := string(bits)
				if content != "" {

					doc, err := dotenv.Parse(content)
					if err == nil {
						for _, node := range doc.ToArray() {
							if node.Type == dotenv.VARIABLE_TOKEN {
								if node.Key == nil {
									continue
								}

								key := *node.Key

								// do not overwrite existing env variables
								if env.Has(key) {
									continue
								}

								/*
									if key == "AZURE_CLIENT_SECRET_KEY" {
										if node.Value == "true" || node.Value == "1" {
											node.Value = "CLIENT_SECRET"
										}

										secretValue, err := getOsSecret(node.Value, "hyprx-secrets-akv")
										if err == nil {
											env.Set("AZURE_CLIENT_SECRET", secretValue)
										}
										continue
									}

									if key == "AZURE_CLIENT_CERTIFICATE_PASSWORD_KEY" {
										if node.Value == "true" || node.Value == "1" {
											node.Value = "CERTIFICATE_PASSWORD"
										}
										secretValue, err := getOsSecret(node.Value, "hyprx-secrets-akv")
										if err == nil {
											env.Set("AZURE_CLIENT_CERTIFICATE_PASSWORD", secretValue)
										}
										continue
									}
								*/

								if strings.HasPrefix(key, "AZURE_") {
									env.Set(key, node.Value)
								}

								env.Set(*node.Key, node.Value)
							}
						}
					}
				}
			}
		}
	}

	useIdentityRaw := env.Get("AZURE_IDENTITY")
	if strings.EqualFold(useIdentityRaw, "true") || strings.EqualFold(useIdentityRaw, "1") {
		clientId := env.Get("AZURE_CLIENT_ID")
		if clientId == "" {
			return azidentity.NewManagedIdentityCredential(nil)
		}

		return azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ClientID(clientId),
		})
	}

	azCliCredential, err3 := azidentity.NewAzureCLICredential(nil)
	if err3 != nil {
		return nil, err3
	}

	credentials := []azcore.TokenCredential{}

	if (env.Has("AZURE_TENANT_ID") && env.Has("AZURE_CLIENT_ID")) && env.Has("AZURE_CLIENT_SECRET") || env.Has("AZURE_CLIENT_CERTIFICATE_PATH") {
		println("Using environment credentials")
		envCredentials, err1 := azidentity.NewEnvironmentCredential(nil)
		if err1 != nil {
			return nil, err1
		}
		credentials = append(credentials, envCredentials)
	}

	if azCliCredential != nil {
		credentials = append(credentials, azCliCredential)
	}

	if interactive != nil && *interactive == "device-code" {
		creds, err := newDeviceCode(ctx)
		if err != nil {
			return nil, err
		}
		if creds != nil {
			credentials = append(credentials, creds)
		} else {
			println("Device code authentication is not supported in this environment.")
			os.Exit(6)
		}
	} else if interactive != nil && *interactive == "interactive" {
		creds, err := newAzInteractive(ctx)
		if err != nil {
			return nil, err
		}
		if creds != nil {
			credentials = append(credentials, creds)
		}

		credentials = append(credentials, creds)
	}

	return azidentity.NewChainedTokenCredential(credentials, nil)
}
