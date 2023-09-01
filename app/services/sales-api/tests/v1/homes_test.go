package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/handlers"
	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
)

// HomeTests holds methods for each home subtest. This type allows passing
// dependencies for test while still providing a convenient synta when
// subtests are registered.
type HomeTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// Test_Homes is the entry point for testing home management apis.
func Test_Homes(t *testing.T) {
	t.Parallel()

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	shutdown := make(chan os.Signal, 1)
	tests := HomeTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		userToken:  test.Token("user@example.com", "gophers"),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	// ------------------------------------------------------------------------
	seed := func(ctx context.Context, usrCore *user.Core, hmeCore *home.Core) ([]user.User, []home.Home, error) {
		usrs, err := usrCore.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding users : %w", err)
		}

		hmes1, err := home.TestGenerateSeedHomes(5, hmeCore, usrs[0].ID)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding homes1 : %w", err)
		}

		hmes2, err := home.TestGenerateSeedHomes(5, hmeCore, usrs[1].ID)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding homes2 : %w", err)
		}

		var hmes []home.Home
		hmes = append(hmes, hmes1...)
		hmes = append(hmes, hmes2...)

		return usrs, hmes, nil
	}

	t.Log("Go seeding ...")

	hmes, _, err := seed(context.Background(), api.User, api.Home)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	t.Run()
}

// TODO finish this test file.
