// Package auditdb contains audit related CRUD functionality.
package auditdb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for audit database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new audit record into the database.
func (s *Store) Create(ctx context.Context, a auditbus.Audit) error {
	const q = `
	INSERT INTO audit
		(id, obj_id, obj_domain, obj_name, actor_id, action, data, message, timestamp)
	VALUES
		(:id, :obj_id, :obj_domain, :obj_name, :actor_id, :action, :data, :message, :timestamp)`

	dbAudit, err := toDBAudit(a)
	if err != nil {
		return err
	}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, dbAudit); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter auditbus.QueryFilter) ([]auditbus.Audit, error) {
	data := map[string]any{}

	const q = `
	SELECT
		id, obj_id, obj_domain, obj_name, actor_id, action, data, message, timestamp
	FROM
		audit`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	buf.WriteString(" ORDER BY timestamp DESC")

	var dbAudits []audit
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbAudits); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusAudits(dbAudits)
}
