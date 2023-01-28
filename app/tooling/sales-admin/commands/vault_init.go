package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/service/foundation/vault"
)

const credentialsFileName = "/vault/credentials.json"

// VaultInit sets up a newly provisioned vault instance.
func VaultInit(vaultConfig vault.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	vaultSrv, err := vault.New(vault.Config{
		Address:   vaultConfig.Address,
		MountPath: vaultConfig.MountPath,
	})
	if err != nil {
		return fmt.Errorf("constructing vault: %w", err)
	}

	// =========================================================================

	log.Println("Check system is already initialized")

	// NOTE: This assumes the Vault POD is never restarted.

	initResponse, err := checkIfCredFileExists()
	if err == nil {
		log.Printf("rootToken: %s", initResponse.RootToken)
		vaultSrv.SetToken(initResponse.RootToken)

		if err = vaultSrv.CheckToken(ctx, vaultConfig.Token); err == nil {
			log.Printf("application token %q exists", vaultConfig.Token)
			return nil
		}
	}

	// =========================================================================

	log.Println("Initializing vault")

	initResponse, err = vaultSrv.SystemInit(ctx, 1, 1)
	if err != nil {
		if errors.Is(err, vault.ErrAlreadyInitialized) {
			return fmt.Errorf("vault is already initialized but we don't have the credentials file")
		}
		return fmt.Errorf("unable to initialize Vault instance: %w", err)
	}

	b, err := json.Marshal(initResponse)
	if err != nil {
		return errors.New("unable to marshal")
	}

	if err := os.WriteFile(credentialsFileName, b, 0644); err != nil {
		return fmt.Errorf("unable to write %s file: %w", credentialsFileName, err)
	}

	log.Printf("rootToken: %s", initResponse.RootToken)
	vaultSrv.SetToken(initResponse.RootToken)

	log.Println("checking token exists")
	err = vaultSrv.CheckToken(ctx, vaultConfig.Token)
	if err == nil {
		log.Printf("token already exists: %s", vaultConfig.Token)
		return nil
	}

	// =========================================================================

	log.Println("Unsealing vault")

	err = vaultSrv.Unseal(ctx, initResponse.KeysB64[0])
	if err != nil {
		if errors.Is(err, vault.ErrBadRequest) {
			return fmt.Errorf("vault is not initialized. Check for old credentials file: %s", credentialsFileName)
		}
		return fmt.Errorf("error unsealing vault: %w", err)
	}

	// =========================================================================

	log.Println("Mounting path in vault")

	vaultSrv.SetToken(initResponse.RootToken)
	if err := vaultSrv.Mount(ctx); err != nil {
		if errors.Is(err, vault.ErrPathInUse) {
			return fmt.Errorf("unable to mount path: %w", err)
		}
		return fmt.Errorf("error unsealing vault: %w", err)
	}

	// =========================================================================

	log.Println("Creating sales-api policy")

	err = vaultSrv.CreatePolicy(ctx, "sales-api", "secret/data/*", []string{"read", "create", "update"})
	if err != nil {
		return fmt.Errorf("unable to create policy: %w", err)
	}

	// =========================================================================

	log.Printf("Generating sales-api token: %s", vaultConfig.Token)

	// We don't currently save the token because we're always going to specify it.
	err = vaultSrv.CreateToken(ctx, vaultConfig.Token, []string{"sales-api"}, "Sales API")
	if err != nil {
		return fmt.Errorf("unable to create token: %w", err)
	}

	return nil
}

// =============================================================================

func checkIfCredFileExists() (vault.SystemInitResponse, error) {
	if _, err := os.Stat(credentialsFileName); err != nil {
		return vault.SystemInitResponse{}, err
	}

	data, err := os.ReadFile(credentialsFileName)
	if err != nil {
		return vault.SystemInitResponse{}, fmt.Errorf("reading %s file: %s", credentialsFileName, err)
	}

	var response vault.SystemInitResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return vault.SystemInitResponse{}, fmt.Errorf("unmarshalling json: %s", err)
	}

	return response, nil
}
