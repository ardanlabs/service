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
	"strings"
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
	MountPath string
	Token     string
	Client    *http.Client
}

// Vault provides support to access Hashicorp's Vault product for keys.
type Vault struct {
	address   string
	token     string
	mountPath string
	client    *http.Client
	mu        sync.RWMutex
	store     map[string]string
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
		store:     make(map[string]string),
	}, nil
}

// SetToken allows the user to change out the token to use on calls.
func (v *Vault) SetToken(token string) {
	v.token = token
}

// =============================================================================

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

// PrivateKeyPEM searches the key store for a given kid and returns
// the private key in pem format.
func (v *Vault) PrivateKeyPEM(kid string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	privatePEM, err := v.retrieveKID(ctx, kid)
	if err != nil {
		return "", fmt.Errorf("kid lookup failed: %w", err)
	}

	return privatePEM, nil
}

// PublicKeyPEM searches the key store for a given kid and returns
// the public key in pem format.
func (v *Vault) PublicKeyPEM(kid string) (string, error) {
	if pem, err := v.keyLookup(kid); err == nil {
		return pem, nil
	}

	privatePEM, err := v.PrivateKeyPEM(kid)
	if err != nil {
		return "", err
	}

	publicPEM, err := toPublicPEM(privatePEM)
	if err != nil {
		return "", err
	}

	v.mu.Lock()
	{
		v.store[kid] = publicPEM
	}
	v.mu.Unlock()

	return publicPEM, nil
}

// =============================================================================

// Error variables for this set of API calls.
var (
	ErrAlreadyInitialized = errors.New("already initalized")
	ErrBadRequest         = errors.New("bad request")
	ErrPathInUse          = errors.New("path in use")
)

// SystemInitResponse represents the response from a system init call.
type SystemInitResponse struct {
	KeysB64   []string `json:"keys_base64"`
	RootToken string   `json:"root_token"`
}

// SystemInit provides support to initialize a vault system for use.
func (v *Vault) SystemInit(ctx context.Context, shares int, threshold int) (SystemInitResponse, error) {
	url := fmt.Sprintf("%s/v1/sys/init", v.address)

	cfg := struct {
		Shares    int `json:"secret_shares"`
		Threshold int `json:"secret_threshold"`
	}{
		Shares:    shares,
		Threshold: threshold,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cfg); err != nil {
		return SystemInitResponse{}, fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &b)
	if err != nil {
		return SystemInitResponse{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "Vault is already initialized") {
			return SystemInitResponse{}, ErrAlreadyInitialized
		}
		return SystemInitResponse{}, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SystemInitResponse{}, fmt.Errorf("status code: %s", resp.Status)
	}

	var response SystemInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return SystemInitResponse{}, fmt.Errorf("json decode: %w", err)
	}

	return response, nil
}

// Unseal does what the unseal command does.
func (v *Vault) Unseal(ctx context.Context, key string) error {
	url := fmt.Sprintf("%s/v1/sys/unseal", v.address)

	cfg := struct {
		Key string `json:"key"`
	}{
		Key: key,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cfg); err != nil {
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
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return ErrBadRequest
		default:
			return fmt.Errorf("status code: %s", resp.Status)
		}
	}

	return nil
}

// Mount accepts a mount point and mounts vault to that point.
func (v *Vault) Mount(ctx context.Context) error {
	mounts, err := v.listMounts(ctx)
	if err != nil {
		return fmt.Errorf("error getting mount list: %w", err)
	}

	// Mount already exists so we'll do nothing.
	if _, ok := mounts[v.mountPath]; ok {
		return nil
	}

	url := fmt.Sprintf("%s/v1/sys/mounts/%s", v.address, v.mountPath)

	cfg := struct {
		Type    string            `json:"type"`
		Options map[string]string `json:"options"`
	}{
		Type:    "kv-v2",
		Options: map[string]string{"version": "2"},
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cfg); err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "path is already in use at") {
			return ErrPathInUse
		}
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("status code: %s", resp.Status)
	}

	return nil
}

// CreatePolicy defines a policy inside of Vault.
func (v *Vault) CreatePolicy(ctx context.Context, name string, path string, capabilities []string) error {
	url := fmt.Sprintf("%s/v1/sys/policies/acl/%s", v.address, name)

	policy := map[string]string{
		"policy": fmt.Sprintf(
			`path %q { capabilities = ["%s"] }`,
			path,
			strings.Join(capabilities, "\",\"")),
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(policy); err != nil {
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %s", resp.Status)
	}

	return nil
}

// CheckToken validates the specified token exists.
func (v *Vault) CheckToken(ctx context.Context, token string) error {
	url := fmt.Sprintf("%s/v1/auth/token/lookup", v.address)

	t := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(t); err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return fmt.Errorf("lookup request: %w", err)
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
		return fmt.Errorf("token doesn't exist: %s", token)
	}

	return nil
}

// CreateToken creates a new token in Vault.
func (v *Vault) CreateToken(ctx context.Context, id string, policies []string, displayName string) error {
	url := fmt.Sprintf("%s/v1/auth/token/create", v.address)

	cfg := struct {
		ID          string   `json:"id,omitempty"`
		Policies    []string `json:"policies,omitempty"`
		DisplayName string   `json:"display_name"`
	}{
		ID:          id,
		Policies:    policies,
		DisplayName: displayName,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cfg); err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
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

// =============================================================================

// keyLookup performs a safe lookup in the store map.
func (v *Vault) keyLookup(kid string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if pem, exists := v.store[kid]; exists {
		return pem, nil
	}

	return "", errors.New("not found")
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

// listMounts returns the set of mount points that exist.
func (v *Vault) listMounts(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/sys/mounts", v.address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %s", resp.Status)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	return response, nil
}

// =============================================================================

// toPublicPEM was taken from the JWT package to reduce the dependency. It
// accepts a PEM encoding of a RSA private key and converts to a PEM encoded
// public key.
func toPublicPEM(privateKeyPEM string) (string, error) {
	var block *pem.Block
	if block, _ = pem.Decode([]byte(privateKeyPEM)); block == nil {
		return "", errors.New("invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key")
	}

	var parsedKey interface{}
	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
	}

	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("key is not a valid RSA private key")
	}

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("marshaling public key: %w", err)
	}

	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	var buf bytes.Buffer
	if err := pem.Encode(&buf, &publicBlock); err != nil {
		return "", fmt.Errorf("encoding to public PEM: %w", err)
	}

	return buf.String(), nil
}
