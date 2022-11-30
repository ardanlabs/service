package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ardanlabs/service/foundation/vault"
)

const credentialsFileName = "/tmp/vault-init.json"

// VaultInit sets up a newly provisioned vault instance.
func VaultInit(vaultConfig vault.Config) error {
	var ctx = context.Background()

	vaultSrv, err := vault.New(vault.Config{
		Address:   vaultConfig.Address,
		MountPath: vaultConfig.MountPath,
	})
	if err != nil {
		return fmt.Errorf("constructing vault: %w", err)
	}

	initResponse, err := ReadCredentialsFile()
	switch {
	case errors.Is(err, os.ErrNotExist):
		log.Println("credential file doesn't exist, initializing vault")

		initResponse, err = vaultSrv.Init(ctx, &vault.InitRequest{
			SecretShares:    1,
			SecretThreshold: 1,
		})
		if err != nil {
			// In order for an initContainer to continue we need to not return an error in this case
			if strings.Contains(err.Error(), "Vault is already initialized") {
				log.Fatalf("Vault is already initialized but we don't have the credentials file")
			}
			log.Fatalf("unable to initialize Vault instance: %v", err.Error())
		}

		b, err := json.Marshal(initResponse)
		if err != nil {
			log.Fatalf("unable to marshal")
		}

		err = os.WriteFile(credentialsFileName, b, 0644)
		if err != nil {
			log.Fatalf("unable to write %s file: %s", credentialsFileName, err)
		}

	default:
		log.Fatalf("unable to read credentials file: %s", err)
	}

	log.Println("Unsealing vault")
	err = vaultSrv.Unseal(ctx, &vault.UnsealOpts{
		Key: initResponse.KeysB64[0],
	})
	if err != nil {
		if strings.Contains(err.Error(), "400 Bad Request") {
			log.Fatalf("Vault is not initialized. Check for old credentials file: %s", credentialsFileName)
		}
		log.Fatalf("Error unsealing vault: %s", err)
	}

	vaultSrv.SetToken(initResponse.RootToken)

	log.Println("Mounting path in vault")
	err = vaultSrv.Mount(ctx, &vault.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "2"},
	})
	if err != nil {
		if !strings.Contains(err.Error(), "path is already in use at") {
			log.Fatalf("Unable to mount path: %s", err)
		}
	}

	return nil
}

func ReadCredentialsFile() (*vault.InitResponse, error) {
	var initResponse = &vault.InitResponse{}

	if _, err := os.Stat(credentialsFileName); err != nil {
		return nil, err
	}

	initResponseJson, err := os.ReadFile(credentialsFileName)
	if err != nil {
		log.Fatalf("Error reading %s file: %s", credentialsFileName, err)
	}

	err = json.Unmarshal(initResponseJson, &initResponse)
	if err != nil {
		log.Fatalf("Error Unmarshalling json: %s", err)
	}

	return initResponse, nil
}
