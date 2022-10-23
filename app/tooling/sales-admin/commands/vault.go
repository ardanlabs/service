package commands

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/ardanlabs/service/foundation/vault"
	"github.com/hashicorp/vault/api"
)

// Vault loads the current private key into the vault system.
func Vault(vaultConfig vault.Config, keysFolder string) error {
	client, err := api.NewClient(&api.Config{
		Address: vaultConfig.Address,
	})
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	client.SetToken(vaultConfig.Token)

	if err := loadKeys(client, vaultConfig.MountPath, vaultConfig.SecretPath, os.DirFS(keysFolder)); err != nil {
		return err
	}

	return nil
}

func loadKeys(client *api.Client, mountPath string, secretPath string, fsys fs.FS) error {
	v2 := client.KVv2(mountPath)

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walkdir failure: %w", err)
		}

		if dirEntry.IsDir() {
			return nil
		}

		if path.Ext(fileName) != ".pem" {
			return nil
		}

		file, err := fsys.Open(fileName)
		if err != nil {
			return fmt.Errorf("opening key file: %w", err)
		}
		defer file.Close()

		// limit PEM file size to 1 megabyte. This should be reasonable for
		// almost any PEM file and prevents shenanigans like linking the file
		// to /dev/random or something like that.
		privatePEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
		if err != nil {
			return fmt.Errorf("reading auth private key: %w", err)
		}

		key := strings.TrimSuffix(dirEntry.Name(), ".pem")
		fmt.Println("Loading Key:", key)

		if _, err := v2.Put(context.Background(), secretPath, map[string]interface{}{key: string(privatePEM)}); err != nil {
			return fmt.Errorf("put: %w", err)
		}

		return nil
	}

	fmt.Print("\n")
	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}
