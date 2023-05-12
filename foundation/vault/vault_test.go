package vault_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/ardanlabs/service/foundation/docker"
	"github.com/ardanlabs/service/foundation/vault"
)

func Test_Vault(t *testing.T) {
	const address = "0.0.0.0:8200"
	const token = "mytoken"
	const mountPath = "secret"
	const key = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
	const image = "hashicorp/vault:1.13"
	const port = "8200"

	args := []string{"-e", "VAULT_DEV_ROOT_TOKEN_ID=" + token, "-e", "VAULT_DEV_LISTEN_ADDRESS=" + address}

	c, err := docker.StartContainer(image, port, args...)
	if err != nil {
		t.Fatalf("starting container: %s", err)
	}
	defer docker.StopContainer(c.ID)

	t.Logf("Image:       %s\n", image)
	t.Logf("ContainerID: %s\n", c.ID)
	t.Logf("Host:        %s\n", c.Host)

	// Give Vault time to initialize.
	time.Sleep(time.Second)

	// -------------------------------------------------------------------------

	vault, err := vault.New(vault.Config{
		Address:   "http://" + c.Host,
		MountPath: mountPath,
		Token:     token,
	})
	if err != nil {
		t.Fatalf("Should be able to construct our Vault API : %s", err)
	}

	pkExp, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Should be able to generate a private key : %s", err)
	}

	pbExp := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pkExp),
	}
	var expPEM bytes.Buffer
	if err := pem.Encode(&expPEM, &pbExp); err != nil {
		t.Fatalf("Should be able to encode pk to PEM : %s", err)
	}

	if err := vault.AddPrivateKey(context.Background(), key, expPEM.Bytes()); err != nil {
		t.Fatalf("Should be able to put the PEM into Vault : %s", err)
	}

	gotPEM, err := vault.PrivateKey(key)
	if err != nil {
		t.Fatalf("Should be able to pull the private key from Vault : %s", err)
	}

	if expPEM.String() != gotPEM {
		t.Logf("got: %s", gotPEM)
		t.Logf("exp: %s", expPEM.String())
		t.Error("Should be able to see the keys match")
	}
}
