package keystore_test

import (
	_ "embed" // Embed all sql documents

	"strings"
	"testing"
	"testing/fstest"

	"github.com/ardanlabs/service/foundation/keystore"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

//go:embed test.pem
var keyDoc []byte

func TestRead(t *testing.T) {
	t.Log("Given the need to parse a directory of private key files.")
	{
		fileName := "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem"
		keyID := strings.TrimRight(fileName, ".pem")
		fsys := fstest.MapFS{}
		fsys[fileName] = &fstest.MapFile{Data: keyDoc}

		testID := 0
		t.Logf("\tTest %d:\tWhen handling a directory of %d file(s).", testID, len(fsys))
		{
			ks, err := keystore.NewFS(fsys)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to construct key store: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to construct key store.", success, testID)

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
