package keystore_test

import (
	"embed" // Calls init function.
	"testing"

	"github.com/ardanlabs/service/foundation/keystore"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

//go:embed *.pem
var keyDocs embed.FS

func TestRead(t *testing.T) {
	t.Log("Given the need to parse a directory of private key files.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a directory of keyfile(s).", testID)
		{
			ks, err := keystore.NewFS(keyDocs)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to construct key store: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to construct key store.", success, testID)

			const keyID = "test"
			pk, err := ks.PrivateKey(keyID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to find key in store: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to find key in store.", success, testID)

			if err := pk.Validate(); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to validate the key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to validate the key.", success, testID)
		}
	}
}
