// Package vault provides support for accessing Hashicorp's vault service
// to access private keys.
package vault

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// This provides a default client configuration, but it's recommended
// this is replaced by the user with application specific settings using
// the WithClient function at the time a GraphQL is constructed.
var defaultClient = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// Config represents the mandatory settings needed to work with Vault.
type Config struct {
	Address   string
	Token     string
	MountPath string
	Client    *http.Client
}

// Vault provides support to access Hashicorp's Vault product for keys.
type Vault struct {
	address   string
	token     string
	mountPath string
	client    *http.Client
	mu        sync.RWMutex
	store     map[string]*rsa.PublicKey
}

// New constructs a vault for use.
func New(cfg Config) (*Vault, error) {
	if cfg.Client == nil {
		cfg.Client = &defaultClient
	}

	return &Vault{
		address:   cfg.Address,
		token:     cfg.Token,
		mountPath: cfg.MountPath,
		client:    cfg.Client,
		store:     make(map[string]*rsa.PublicKey),
	}, nil
}

// AddPrivateKey adds a new private key into vault as PEM encoded.
func (v *Vault) AddPrivateKey(ctx context.Context, kid string, pem []byte) error {
	url := fmt.Sprintf("%s/v1/%s/data/%s", v.address, v.mountPath, kid)

	data := struct {
		M map[string]string `json:"data"`
	}{
		M: map[string]string{
			"pem": string(pem),
		},
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(data); err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &b)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %s", resp.Status)
	}

	return nil
}

// PrivateKey searches the key store for a given kid and returns
// the private key.
func (v *Vault) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	privatePEM, err := v.retrieveKID(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("kid lookup failed: %w", err)
	}

	privateKey, err := toPrivateKey(privatePEM)
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

// =============================================================================

// keyLookup performs a safe lookup in the store map.
func (v *Vault) keyLookup(kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if pk, exists := v.store[kid]; exists {
		return pk, nil
	}

	return nil, errors.New("not found")
}

// retrieveKID performs the HTTP call against the Vault service for the
// specified kid and returns the pem value.
func (v *Vault) retrieveKID(ctx context.Context, kid string) (string, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", v.address, v.mountPath, kid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %s", resp.Status)
	}

	var data struct {
		RequestID     string `json:"request_id"`
		LeaseID       string `json:"lease_id"`
		Renewable     bool   `json:"renewable"`
		LeaseDuration int    `json:"lease_duration"`
		Data          struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decoding: %w", err)
	}

	pem, ok := data.Data.Data["pem"]
	if !ok {
		return "", fmt.Errorf("kid %q does not exist", kid)
	}

	return pem, nil
}

// =============================================================================

// toPrivateKey was taken from the JWT package to reduce the dependency. It
// accepts a PEM encoding of a RSA private key and converts to a proper type.
func toPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	var block *pem.Block
	if block, _ = pem.Decode([]byte(privateKeyPEM)); block == nil {
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
