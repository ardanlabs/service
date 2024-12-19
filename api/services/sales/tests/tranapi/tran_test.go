package tran_test

import (
	"testing"

	"github.com/ardanlabs/service/app/sdk/apitest"
)

func Test_Tran(t *testing.T) {
	t.Parallel()

	test := apitest.New(t, "Test_VProduct")

	// -------------------------------------------------------------------------

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
}
