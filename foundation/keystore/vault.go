package keystore

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

// VaultConfig represents the mandatory settings needed to work with Vault.
type VaultConfig struct {
	Address    string
	Token      string
	MountPath  string
	SecretPath string
}

// Vault provides support to access Hashicorp's Vault product for keys.
type Vault struct {
	address    string
	mountPath  string
	secretPath string
	client     *api.Client
	mu         sync.RWMutex
	store      map[string]*rsa.PublicKey
}

// NewVault constructs a vault for use.
func NewVault(cfg VaultConfig) (*Vault, error) {
	client, err := api.NewClient(&api.Config{
		Address: cfg.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}
	client.SetToken(cfg.Token)

	return &Vault{
		address:    cfg.Address,
		mountPath:  cfg.MountPath,
		secretPath: cfg.SecretPath,
		client:     client,
		store:      make(map[string]*rsa.PublicKey),
	}, nil
}

// PrivateKey searches the key store for a given kid and returns
// the private key.
func (v *Vault) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := v.client.KVv2(v.mountPath).Get(ctx, v.secretPath)
	if err != nil {
		return nil, fmt.Errorf("kid lookup failed: %w", err)
	}

	privatePEM, exists := data.Data[kid]
	if !exists {
		return nil, errors.New("kid not found")
	}

	key, ok := privatePEM.(string)
	if !ok {
		return nil, errors.New("private PEM encoding is wrong")
	}

	privateKey, err := parseRSAPrivateKeyFromPEM([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	return privateKey, nil
}

// PublicKey searches the key store for a given kid and returns
// the public key.
func (v *Vault) PublicKey(kid string) (*rsa.PublicKey, error) {
	if pk, err := v.keyLookup(kid); err == nil {
		return pk, nil
	}

	privateKey, err := v.PrivateKey(kid)
	if err != nil {
		return nil, err
	}

	v.mu.Lock()
	{
		v.store[kid] = &privateKey.PublicKey
	}
	v.mu.Unlock()

	return &privateKey.PublicKey, nil
}

// keyLookup performs a safe lookup in the store map.
func (v *Vault) keyLookup(kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if pk, exists := v.store[kid]; exists {
		return pk, nil
	}

	return nil, errors.New("not found")
}
