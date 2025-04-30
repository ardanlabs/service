// Package auditdb contains audit related CRUD functionality.
package auditdb

import (
	"bytes"
	"context"
	"fmt"
	"strings"

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
		(id, primary_id, user_id, action, data, message, timestamp)
	VALUES
		(:id, :primary_id, :user_id, :action, :data, :message, :timestamp)`

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
		id, primary_id, user_id, action, data, message, timestamp
	FROM
		audit`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	switch {
	case !strings.Contains(buf.String(), "WHERE"):
		buf.WriteString(" WHERE org_id = :org_id")

	default:
		buf.WriteString(" AND org_id = :org_id")
	}

	buf.WriteString(" ORDER BY timestamp DESC")

	var dbAudits []audit
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbAudits); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusAudits(dbAudits)
}
