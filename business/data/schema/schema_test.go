package schema

import (
	_ "embed"

	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestParse(t *testing.T) {
	t.Log("Given the need to parse a sql migration file.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling the embedded schema.", testID)
		{
			migs := parseMigrations(schemaDoc)
			var buf bytes.Buffer
			for _, mig := range migs {
				buf.WriteString(fmt.Sprintf("-- version: %.1f\n", mig.Version))
				buf.WriteString(fmt.Sprintf("-- description: %s\n", mig.Description))
				buf.WriteString(mig.Script)
			}

			sql := strings.ToLower(schemaDoc)
			if sql != buf.String() {
				t.Logf("got: %v", buf.Bytes())
				t.Logf("exp: %v", []byte(sql))
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse migrations.", failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse migrations.", success, testID)
		}
	}
}
