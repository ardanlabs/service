package authapi

import (
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/google/uuid"
)

// Error represents an error in the system.
type Error struct {
	Message string `json:"message"`
}

// Error implements the error interface.
func (err Error) Error() string {
	return err.Message
}

// AuthInfo defines the information required to perform an authorization.
type AuthInfo struct {
	Claims auth.Claims
	UserID uuid.UUID
	Rule   string
}

type AuthResp struct {
	UserID uuid.UUID
	Claims auth.Claims
}
