package keystore

import (
	"os"
	"testing"
)

func TestBill(t *testing.T) {
	ks, err := Read(os.DirFS("/Users/bill/code/go/src/github.com/ardanlabs/service/zarf/keys"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ks)
}
