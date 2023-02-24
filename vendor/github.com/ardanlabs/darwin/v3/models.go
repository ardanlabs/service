package darwin

import (
	"crypto/md5"
	"fmt"
	"time"
)

const (
	// Ignored means that the migrations was not appied to the database.
	Ignored Status = iota

	// Applied means that the migrations was successfully applied to the database.
	Applied

	// Pending means that the migrations is a new migration and it is waiting
	// to be applied to the database.
	Pending

	// Error means that the migration could not be applied to the database.
	Error
)

// =============================================================================

// Status is a migration status value.
type Status int

// String implements the Stringer interface.
func (s Status) String() string {
	switch s {
	case Ignored:
		return "IGNORED"
	case Applied:
		return "APPLIED"
	case Pending:
		return "PENDING"
	case Error:
		return "ERROR"
	default:
		return "INVALID"
	}
}

// =============================================================================

// Migration represents a database migrations.
type Migration struct {
	Version     float64
	Description string
	Script      string
}

// Checksum calculate the Script md5.
func (m Migration) Checksum() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(m.Script)))
}

// MigrationInfo is a struct used in the infoChan to inform clients about
// the migration being applied.
type MigrationInfo struct {
	Status    Status
	Error     error
	Migration Migration
}

// MigrationRecord is the entry in schema table.
type MigrationRecord struct {
	Version       float64
	Description   string
	Checksum      string
	AppliedAt     time.Time
	ExecutionTime time.Duration
}

type byMigrationRecordVersion []MigrationRecord

func (b byMigrationRecordVersion) Len() int           { return len(b) }
func (b byMigrationRecordVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMigrationRecordVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
