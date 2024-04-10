package authapi

import (
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/google/uuid"
)

// AuthInfo defines the information required to perform an authorization.
type AuthInfo struct {
	Claims auth.Claims
	UserID uuid.UUID
	Rule   string
}

// Token represents the response for the token call.
type Token struct {
	Token string `json:"token"`
}
