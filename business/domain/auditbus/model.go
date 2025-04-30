package auditbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Audit represents information about an individual audit record.
type Audit struct {
	ID        uuid.UUID
	PrimaryID uuid.UUID
	UserID    uuid.UUID
	Action    string
	Data      json.RawMessage
	Message   string
	Timestamp time.Time
}
