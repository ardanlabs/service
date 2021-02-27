package schema

import (
	_ "embed"

	"bytes"
	"fmt"
	"strings"
	"testing"
)

//go:embed schema.sql
var sqlTest string

func TestParse(t *testing.T) {
	t.Log("Given the need to parse a sql migration file.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a properly formed file.", testID)
		{
			migs := parseMigrations(sqlTest)
			var buf bytes.Buffer
			for _, mig := range migs {
				buf.WriteString(fmt.Sprintf("-- version: %.1f\n", mig.Version))
				buf.WriteString(fmt.Sprintf("-- description: %s\n", mig.Description))
				buf.WriteString(mig.Script)
			}

			if strings.ToLower(sqlTest) != buf.String() {
				t.Logf("got: %v", buf.Bytes())
				t.Logf("exp: %v", []byte(strings.ToLower(sqlTest)))
				t.Fatal("NOT THE SAME")
			}
		}
	}
}
