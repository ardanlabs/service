package tran_test

import (
	"fmt"
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/api/http/api/apitest"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	var err error

	c, err = dbtest.StartDB()
	if err != nil {
		return 1, err
	}

	return m.Run(), nil
}

// =============================================================================

func Test_Tran(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	test := apitest.StartTest(t, c, "Test_VProduct")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, create200(sd), "query-200")
}
