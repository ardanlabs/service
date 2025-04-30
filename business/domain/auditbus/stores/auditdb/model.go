package auditdb

import (
	"encoding/json"
	"time"

	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

type audit struct {
	ID        uuid.UUID          `db:"id"`
	PrimaryID uuid.UUID          `db:"primary_id"`
	UserID    uuid.UUID          `db:"user_id"`
	Action    string             `db:"action"`
	Data      types.NullJSONText `db:"data"`
	Message   string             `db:"message"`
	Timestamp time.Time          `db:"timestamp"`
}

func toDBAudit(bus auditbus.Audit) (audit, error) {
	db := audit{
		ID:        bus.ID,
		PrimaryID: bus.PrimaryID,
		UserID:    bus.UserID,
		Action:    bus.Action,
		Data:      types.NullJSONText{JSONText: []byte(bus.Data), Valid: true},
		Message:   bus.Message,
		Timestamp: bus.Timestamp,
	}

	return db, nil
}

func toBusAudit(db audit) (auditbus.Audit, error) {
	bus := auditbus.Audit{
		ID:        db.ID,
		PrimaryID: db.PrimaryID,
		UserID:    db.UserID,
		Action:    db.Action,
		Data:      json.RawMessage(db.Data.JSONText),
		Message:   db.Message,
		Timestamp: db.Timestamp,
	}

	return bus, nil
}

func toBusAudits(dbs []audit) ([]auditbus.Audit, error) {
	audits := make([]auditbus.Audit, len(dbs))

	for i, db := range dbs {
		a, err := toBusAudit(db)
		if err != nil {
			return nil, err
		}

		audits[i] = a
	}

	return audits, nil
}
