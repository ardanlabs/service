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
