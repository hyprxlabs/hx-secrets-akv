package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// this example shows file storage but any form of byte storage would work
func retrieveRecord() (azidentity.AuthenticationRecord, error) {
	record := azidentity.AuthenticationRecord{}
	authRecordPath := getRecordPath()
	if authRecordPath == "" {
		return record, nil
	}

	if _, err := os.Stat(authRecordPath); os.IsNotExist(err) {
		return record, nil
	}

	b, err := os.ReadFile(authRecordPath)
	if err == nil {
		err = json.Unmarshal(b, &record)
	}
	return record, err
}

func storeRecord(record azidentity.AuthenticationRecord) error {
	authRecordPath := getRecordPath()
	if authRecordPath == "" {
		return nil
	}

	dir := filepath.Dir(authRecordPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	b, err := json.Marshal(record)
	if err == nil {
		err = os.WriteFile(authRecordPath, b, 0600)
	}
	return err
}

func getRecordPath() string {
	targetDir := homeConfigDir()
	if targetDir == "" {
		return ""
	}

	return filepath.Join(targetDir, "credential.cache.json")
}

func newAzInteractive(ctx context.Context) (*azidentity.InteractiveBrowserCredential, error) {
	record, err := retrieveRecord()
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		// If record is zero, the credential will start with no user logged in
		AuthenticationRecord: record,
		// Credentials cache in memory by default. Setting Cache with a
		// nonzero value from cache.New() enables persistent caching.
	})
	if err != nil {
		return nil, err
	}

	if record == (azidentity.AuthenticationRecord{}) {
		// No stored record; call Authenticate to acquire one.
		// This will prompt the user to authenticate interactively.
		record, err = cred.Authenticate(ctx, nil)
		if err != nil {
			return nil, err
		}
		err = storeRecord(record)
		if err != nil {
			return nil, err
		}
	}
	return cred, nil
}

func newDeviceCode(ctx context.Context) (*azidentity.DeviceCodeCredential, error) {
	record, err := retrieveRecord()
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		// If record is zero, the credential will start with no user logged in
		AuthenticationRecord: record,
		// Credentials cache in memory by default. Setting Cache with a
		// nonzero value from cache.New() enables persistent caching.
	})
	if err != nil {
		return nil, err
	}

	if record == (azidentity.AuthenticationRecord{}) {
		// No stored record; call Authenticate to acquire one.
		record, err = cred.Authenticate(ctx, nil)
		if err != nil {
			return nil, err
		}
		err = storeRecord(record)
		if err != nil {
			return nil, err
		}
	}
	return cred, nil
}
