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

const credentialsFileName = "/vault/credentials.json"

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
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			log.Println("credential file doesn't exist, initializing vault")

			initResponse, err = vaultSrv.Init(ctx, &vault.InitRequest{
				SecretShares:    1,
				SecretThreshold: 1,
			})
			if err != nil {
				// In order for an initContainer to continue we need to not return an error in this case.
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
	}

	// =============================================================================================================

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

	// =============================================================================================================

	log.Println("Mounting path in vault")

	vaultSrv.SetToken(initResponse.RootToken)
	err = vaultSrv.Mount(ctx, &vault.MountInput{
		Type:    "kv-v2",
		Options: map[string]string{"version": "2"},
	})
	if err != nil {
		if !strings.Contains(err.Error(), "path is already in use at") {
			log.Fatalf("Unable to mount path: %s", err)
		}
	}

	// =============================================================================================================

	log.Println("Creating sales-api policy")

	policy := `{"policy":"path \"secret/data/*\" {\n  capabilities = [\"read\",\"create\",\"update\"]\n}\n"}`

	err = vaultSrv.CreatePolicy(ctx, "sales-api", policy)
	if err != nil {
		log.Fatalf("Unable to create policy: %s", err)
	}

	// =============================================================================================================

	log.Printf("Generating sales-api token: %s", vaultConfig.Token)

	createRequest := vault.TokenCreateRequest{
		ID:          vaultConfig.Token,
		Policies:    []string{"sales-api"},
		DisplayName: "Sales API",
	}

	// First let's check if it exists already.
	err = vaultSrv.CheckToken(ctx, createRequest)
	if err == nil {
		log.Printf("token already exists: %s", vaultConfig.Token)
		return nil
	}

	// We don't currently save the token because we're always going to specify it.
	err = vaultSrv.CreateToken(ctx, createRequest)
	if err != nil {
		log.Fatalf("Unable to create token: %s", err)
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
