// Package authclient holds the authentication client relevant models and interfaces
package authclient

import (
	"context"
	"encoding/json"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/google/uuid"
)

type Authenticator interface {
	Authenticate(ctx context.Context, authorization string) (AuthenticateResp, error)
	Authorize(ctx context.Context, auth Authorize) error
	Close() error
}

// Authorize defines the information required to perform an authorization.
type Authorize struct {
	UserID uuid.UUID
	Claims auth.Claims
	Rule   string
}

// Decode implements the decoder interface.
func (a *Authorize) Decode(data []byte) error {
	return json.Unmarshal(data, a)
}

// AuthenticateResp defines the information that will be received on authenticate.
type AuthenticateResp struct {
	UserID uuid.UUID
	Claims auth.Claims
}

// Encode implements the encoder interface.
func (ar AuthenticateResp) Encode() ([]byte, string, error) {
	data, err := json.Marshal(ar)
	return data, "application/json", err
}
