package authclient

import (
	"encoding/json"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/google/uuid"
)

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
