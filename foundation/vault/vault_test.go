package vault_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509" // Calls init function.
	"encoding/pem"
	"testing"
	"time"

	"github.com/ardanlabs/service/foundation/docker"
	"github.com/ardanlabs/service/foundation/vault"
	"github.com/hashicorp/vault/api"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func Test_Vault(t *testing.T) {
	const address = "0.0.0.0:8200"
	const token = "myroot"
	const mountPath = "secret"
	const secretPath = "sales"
	const image = "hashicorp/vault:1.12"
	const port = "8200"
	const key = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

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

	t.Log("Given the to talk to Vault for key support.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single key.", testID)
		{
			cfg := api.Config{
				Address: "http://" + c.Host,
			}
			client, err := api.NewClient(&cfg)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a client: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a client.", success, testID)

			client.SetToken(token)
			v2 := client.KVv2(mountPath)

			pkExp, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a private key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a private key.", success, testID)

			pbExp := pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(pkExp),
			}
			var bExp bytes.Buffer
			if err := pem.Encode(&bExp, &pbExp); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to encode pk to PEM: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to encode pk to PEM.", success, testID)

			if _, err := v2.Put(context.Background(), secretPath, map[string]interface{}{key: bExp.String()}); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to put the PEM into Vault: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to put the PEM into Vault.", success, testID)

			vault, err := vault.New(vault.Config{
				Address:    cfg.Address,
				Token:      token,
				MountPath:  mountPath,
				SecretPath: secretPath,
			})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to construct our Vault API: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to construct our Vault API.", success, testID)

			pkGot, err := vault.PrivateKey(key)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to pull the private key from Vault: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to pull the private key from Vault.", success, testID)

			pbGot := pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(pkGot),
			}
			var bGot bytes.Buffer
			if err := pem.Encode(&bGot, &pbGot); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to encode the returned private key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to encode the returned private key.", success, testID)

			if bExp.String() != bGot.String() {
				t.Logf("\t\tTest %d:\texp: %s", testID, bExp.String())
				t.Logf("\t\tTest %d:\tgot: %s", testID, bGot.String())
				t.Fatalf("\t%s\tTest %d:\tShould be able to see the keys match: %v", failed, testID, err)
			}
		}
	}
}
