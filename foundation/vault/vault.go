// Package vault provides support for accessing Hashicorp's vault service
// to access private keys.
package vault

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

// Config represents the mandatory settings needed to work with Vault.
type Config struct {
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

// New constructs a vault for use.
func New(cfg Config) (*Vault, error) {
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

// =============================================================================

// parseRSAPrivateKeyFromPEM was taken from the JWT package to reduce the dependency.
func parseRSAPrivateKeyFromPEM(key []byte) (*rsa.PrivateKey, error) {
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, errors.New("invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key")
	}

	var parsedKey interface{}
	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	pkey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not a valid RSA private key")
	}

	return pkey, nil
}
