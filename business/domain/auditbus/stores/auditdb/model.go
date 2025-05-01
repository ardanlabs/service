package auditdb

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/ardanlabs/service/business/types/domain"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

type audit struct {
	ID        uuid.UUID          `db:"id"`
	ObjID     uuid.UUID          `db:"obj_id"`
	ObjDomain string             `db:"obj_domain"`
	ObjName   string             `db:"obj_name"`
	ActorID   uuid.UUID          `db:"actor_id"`
	Action    string             `db:"action"`
	Data      types.NullJSONText `db:"data"`
	Message   string             `db:"message"`
	Timestamp time.Time          `db:"timestamp"`
}

func toDBAudit(bus auditbus.Audit) (audit, error) {
	db := audit{
		ID:        bus.ID,
		ObjID:     bus.ObjID,
		ObjDomain: bus.ObjDomain.String(),
		ObjName:   bus.ObjName.String(),
		ActorID:   bus.ActorID,
		Action:    bus.Action,
		Data:      types.NullJSONText{JSONText: []byte(bus.Data), Valid: true},
		Message:   bus.Message,
		Timestamp: bus.Timestamp,
	}

	return db, nil
}

func toBusAudit(db audit) (auditbus.Audit, error) {
	domain, err := domain.Parse(db.ObjDomain)
	if err != nil {
		return auditbus.Audit{}, fmt.Errorf("parse domain: %w", err)
	}

	name, err := name.Parse(db.ObjName)
	if err != nil {
		return auditbus.Audit{}, fmt.Errorf("parse name: %w", err)
	}

	bus := auditbus.Audit{
		ID:        db.ID,
		ObjID:     db.ObjID,
		ObjDomain: domain,
		ObjName:   name,
		ActorID:   db.ActorID,
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
