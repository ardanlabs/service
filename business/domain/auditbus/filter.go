package auditbus

import (
	"time"

	"github.com/ardanlabs/service/business/types/domain"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ObjID     *uuid.UUID
	ObjDomain *domain.Domain
	ActorID   *uuid.UUID
	Action    *string
	Since     *time.Time
	Until     *time.Time
}
