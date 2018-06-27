package tests

import (
	"os"
	"testing"

	"github.com/ardanlabs/service/cmd/crud/handlers"
	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/platform/web"
)

var a *web.App
var test *tests.Test

// TestMain is the entry point for testing.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	test = tests.New()
	defer test.TearDown()
	a = handlers.API(test.Log, test.MasterDB).(*web.App)
	return m.Run()
}
