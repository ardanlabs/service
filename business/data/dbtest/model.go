package dbtest

import (
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
)

// User represents an app user specified for the test.
type User struct {
	user.User
	Token    string
	Products []product.Product
	Homes    []home.Home
}

// SeedData represents data that was seeded for the test.
type SeedData struct {
	Users  []User
	Admins []User
}
