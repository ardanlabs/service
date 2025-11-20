package homebus

import (
	"time"

	"github.com/ardanlabs/service/business/types/home"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID               *uuid.UUID
	UserID           *uuid.UUID
	Type             *home.Home
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
}
