package auditapp

import (
	"encoding/json"
	"time"

	"github.com/ardanlabs/service/business/domain/auditbus"
)

// Audit represents information about an individual audit record.
type Audit struct {
	ID        string `json:"id"`
	ObjID     string `json:"obj_id"`
	ObjDomain string `json:"obj_domain"`
	ObjName   string `json:"obj_name"`
	ActorID   string `json:"actor_id"`
	Action    string `json:"action"`
	Data      string `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// Encode implements the encoder interface.
func (app Audit) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppAudit(bus auditbus.Audit) Audit {
	return Audit{
		ID:        bus.ID.String(),
		ObjID:     bus.ObjID.String(),
		ObjDomain: bus.ObjDomain.String(),
		ObjName:   bus.ObjName.String(),
		ActorID:   bus.ActorID.String(),
		Action:    bus.Action,
		Data:      string(bus.Data),
		Message:   bus.Message,
		Timestamp: bus.Timestamp.Format(time.RFC3339),
	}
}

func toAppAudits(audits []auditbus.Audit) []Audit {
	app := make([]Audit, len(audits))
	for i, adt := range audits {
		app[i] = toAppAudit(adt)
	}

	return app
}
