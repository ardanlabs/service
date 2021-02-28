// Package keystore implements the auth.KeyStore interface. This implements
// an in-memory keystore for JWT support.
package keystore

import (
	"crypto/rsa"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/pkg/errors"
)

// KeyStore represents an in memory store implementation of the
// KeyStorer interface for use with the auth package.
type KeyStore struct {
	mu    sync.RWMutex
	store map[string]*rsa.PrivateKey
}

// NewMap is given a pre-configured KeyStore as a starting point.
func NewMap(store map[string]*rsa.PrivateKey) *KeyStore {
	return &KeyStore{
		store: store,
	}
}

// NewFS is given a file system rooted inside of a directory that should
// contain private keys files where each file is named for the unique kid
// for that key. Example: 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
func NewFS(fsys fs.FS) (*KeyStore, error) {
	ks := KeyStore{
		store: make(map[string]*rsa.PrivateKey),
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "walkdir failure")
		}

		if dirEntry.IsDir() {
			return nil
		}

		if path.Ext(dirEntry.Name()) != ".pem" {
			return nil
		}

		file, err := fsys.Open(fileName)
		if err != nil {
			return errors.Wrap(err, "open key file")
		}

		privatePEM, err := io.ReadAll(file)
		if err != nil {
			return errors.Wrap(err, "reading auth private key")
		}

		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
		if err != nil {
			return errors.Wrap(err, "parsing auth private key")
		}

		ks.store[strings.TrimRight(dirEntry.Name(), ".pem")] = privateKey
		return nil
	}

	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		return nil, errors.Wrap(err, "walking directory")
	}

	return &ks, nil
}

// Add adds a private key and combination kid to the store.
func (ks *KeyStore) Add(privateKey *rsa.PrivateKey, kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.store[kid] = privateKey
}

// Remove removes a private key and combination kid to the store.
func (ks *KeyStore) Remove(kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	delete(ks.store, kid)
}

// LookupPrivate searches the key store for a given kid.
func (ks *KeyStore) LookupPrivate(kid string) (*rsa.PrivateKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	privateKey, found := ks.store[kid]
	if !found {
		return nil, errors.New("kid lookup failed")
	}
	return privateKey, nil
}

// LookupPublic searches the key store for a given kid.
func (ks *KeyStore) LookupPublic(kid string) (*rsa.PublicKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	privateKey, found := ks.store[kid]
	if !found {
		return nil, errors.New("kid lookup failed")
	}
	return &privateKey.PublicKey, nil
}
