package cmd

import (
	"errors"

	"github.com/99designs/keyring"
)

func Remove(account string, service string) error {
	if service == "" {
		service = "hyprx-secrets-akv"
	}

	if account == "" {
		return errors.New("no account name provided")
	}

	config := keyring.Config{
		ServiceName: service,
		AllowedBackends: []keyring.BackendType{
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
		},
	}

	secretCreds, err := keyring.Open(config)
	if err != nil {
		return err
	}

	return secretCreds.Remove(account)
}

func setOsSecret(account string, service string, secret string) error {
	if service == "" {
		service = "hyprx-secrets-akv"
	}
	if account == "" {
		return errors.New("no account name provided")
	}

	config := keyring.Config{
		ServiceName: service,
		AllowedBackends: []keyring.BackendType{
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
		},
	}
	secretCreds, err := keyring.Open(config)
	if err != nil {
		return err
	}

	return secretCreds.Set(keyring.Item{
		Key:  account,
		Data: []byte(secret),
	})
}

func getOsSecret(account string, service string) (string, error) {
	if service == "" {
		service = "hyprx-secrets-akv"
	}
	if account == "" {
		return "", errors.New("no account name provided")
	}

	config := keyring.Config{
		ServiceName: service,
		AllowedBackends: []keyring.BackendType{
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
		},
	}

	secretCreds, err := keyring.Open(config)
	if err != nil {
		return "", err
	}

	secret, err := secretCreds.Get(account)
	if err != nil {
		return "", err
	}

	return string(secret.Data), nil
}
