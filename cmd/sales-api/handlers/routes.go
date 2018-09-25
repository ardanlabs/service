package handlers

import (
	"crypto/rsa"
	"log"
	"net/http"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(log *log.Logger, masterDB *db.DB, userAuth UserAuth) http.Handler {

	authKeyFunc := func(keyID string) (*rsa.PublicKey, error) {
		if keyID != userAuth.KeyID {
			// TODO(jlw) What do we do about this? Do we want to support rolling keys etc?
		}

		// TODO(jlw) Do we need to explicitly pass a public key in from the config? Since we already have the private key it seems we can just compute the public key when needed.
		key := userAuth.Key.Public().(*rsa.PublicKey)
		return key, nil
	}
	authmw := mid.Auth{
		Parser: auth.NewParser(authKeyFunc, []string{userAuth.Alg}),
	}

	app := web.New(log, mid.RequestLogger, mid.Metrics, mid.ErrorHandler)

	// TODO(jlw) all of our endpoitns require the authmw.Authenticate middleware except for 2. Can we "clone" app and make a second authenticated app or should we do what we're doing below and add Authenticate to almost every route. If we somehow cloned them how would we put them together into one app at the end to return?

	// Register health check endpoint. This route is not authenticated.
	h := Health{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/health", h.Check)

	// Register user management and authentication endpoints.
	u := User{
		MasterDB: masterDB,
		Auth:     userAuth,
	}

	app.Handle("GET", "/v1/users", u.List, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/users", u.Create, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users/:id", u.Retrieve, authmw.Authenticate)
	app.Handle("PUT", "/v1/users/:id", u.Update, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/users/:id", u.Delete, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))

	// This route is not authenticated
	app.Handle("GET", "/v1/users/token", u.Token)

	// Register product and sale endpoints.
	p := Product{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/products", p.List, authmw.Authenticate)
	app.Handle("POST", "/v1/products", p.Create, authmw.Authenticate)
	app.Handle("GET", "/v1/products/:id", p.Retrieve, authmw.Authenticate)
	app.Handle("PUT", "/v1/products/:id", p.Update, authmw.Authenticate)
	app.Handle("DELETE", "/v1/products/:id", p.Delete, authmw.Authenticate)

	return app
}
