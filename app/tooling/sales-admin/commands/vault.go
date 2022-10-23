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
)

// Vault loads the current private key into the vault system.
func Vault(vaultConfig vault.Config, keysFolder string) error {
	vault, err := vault.New(vault.Config{
		Address:   vaultConfig.Address,
		Token:     vaultConfig.Token,
		MountPath: vaultConfig.MountPath,
	})
	if err != nil {
		return fmt.Errorf("constructing vault: %w", err)
	}

	if err := loadKeys(vault, vaultConfig.MountPath, os.DirFS(keysFolder)); err != nil {
		return err
	}

	return nil
}

func loadKeys(vault *vault.Vault, mountPath string, fsys fs.FS) error {
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

		if err := vault.PutKey(context.Background(), key, string(privatePEM)); err != nil {
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
