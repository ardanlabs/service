// Package keystore implements the auth.KeyStore interface. This implements
// an in-memory keystore for JWT support.
package keystore

import (
	"crypto/rsa"
	"io/fs"
	"os"
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

// New is given a pre-configured KeyStore as a starting point.
func New(store map[string]*rsa.PrivateKey) *KeyStore {
	return &KeyStore{
		store: store,
	}
}

// Read is given a directory that contains a collection of pem files
// that represent private keys, where each file is named for the unique kid for
// that key. Example: 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
func Read(directory string, dirReader fs.ReadDirFS, fileReader fs.FS) (*KeyStore, error) {
	directory = strings.TrimRight(directory, "/") + "/"

	dirEntries, err := dirReader.ReadDir(directory)
	if err != nil {
		return nil, errors.New("unable to read directory entries")
	}

	ks := KeyStore{
		store: make(map[string]*rsa.PrivateKey),
	}

	for _, dirEntry := range dirEntries {
		privatePEM, err := fs.ReadFile(fileReader, directory+dirEntry.Name())
		if err != nil {
			return nil, errors.Wrap(err, "reading auth private key")
		}
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
		if err != nil {
			return nil, errors.Wrap(err, "parsing auth private key")
		}
		ks.store[strings.TrimRight(dirEntry.Name(), ".pem")] = privateKey
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

// =============================================================================

// FS implements the fs.FS interface for opening files.
type FS struct{}

// Open implements the fs.FS interface for accessing disk.
func (FS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

// ReadDirFS implements the fs.ReadDirFS interface for listing files.
type ReadDirFS struct {
	FS
}

// ReadDir implements the ReadDirFS for accessing disk.
func (ReadDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}
