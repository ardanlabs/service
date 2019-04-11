package mid

import (
	"github.com/ardanlabs/service/internal/platform/auth"
)

// Middleware holds the required state for all web.Middleware functions in this
// package. Its methods are defined in separate files.
type Middleware struct {
	Authenticator *auth.Authenticator
}
