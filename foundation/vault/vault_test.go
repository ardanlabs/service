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

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
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

	t.Log("Given the need to talk to Vault for key support.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single key.", testID)
		{
			vault, err := vault.New(vault.Config{
				Address:   "http://" + c.Host,
				MountPath: mountPath,
				Token:     token,
			})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to construct our Vault API: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to construct our Vault API.", success, testID)

			pkExp, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a private key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a private key.", success, testID)

			pbExp := pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(pkExp),
			}
			var expPEM bytes.Buffer
			if err := pem.Encode(&expPEM, &pbExp); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to encode pk to PEM: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to encode pk to PEM.", success, testID)

			if err := vault.AddPrivateKey(context.Background(), key, expPEM.Bytes()); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to put the PEM into Vault: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to put the PEM into Vault.", success, testID)

			gotPEM, err := vault.PrivateKey(key)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to pull the private key from Vault: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to pull the private key from Vault.", success, testID)

			if expPEM.String() != gotPEM {
				t.Logf("\t\tTest %d:\texp: %s", testID, expPEM.String())
				t.Logf("\t\tTest %d:\tgot: %s", testID, gotPEM)
				t.Fatalf("\t%s\tTest %d:\tShould be able to see the keys match: %v", failed, testID, err)
			}
		}
	}
}
