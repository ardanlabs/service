package darwin

import (
	"sort"
	"time"
)

// Driver is a database driver abstraction.
type Driver interface {
	Create() error
	Insert(MigrationRecord) error
	UpdateChecksum(checksum string, version float64) error
	All() ([]MigrationRecord, error)
	Exec(string) (time.Duration, error)
}

// =============================================================================

// Darwin is a helper struct to access the Validate and migration functions.
type Darwin struct {
	driver     Driver
	migrations []Migration
}

// New returns a new Darwin struct
func New(driver Driver, migrations []Migration) Darwin {
	return Darwin{
		driver:     driver,
		migrations: migrations,
	}
}

// Validate if the database migrations are applied and consistent.
func (d *Darwin) Validate() error {
	sort.Sort(byMigrationVersion(d.migrations))

	if version, invalid := d.isInvalidVersion(); invalid {
		return &IllegalMigrationVersionError{Version: version}
	}

	if version, dup := d.isDuplicated(); dup {
		return &DuplicateMigrationVersionError{Version: version}
	}

	applied, err := d.driver.All()
	if err != nil {
		return err
	}

	if version, removed := d.wasRemovedMigration(applied); removed {
		return &RemovedMigrationError{Version: version}
	}

	if version, invalid := d.isInvalidChecksumMigration(applied); invalid {
		return &InvalidChecksumError{Version: version}
	}

	return nil
}

// Migrate executes the missing migrations in database.
func (d *Darwin) Migrate() error {
	if err := d.driver.Create(); err != nil {
		return err
	}

	if err := d.Validate(); err != nil {
		return err
	}

	planned, err := d.planMigration()
	if err != nil {
		return err
	}

	for _, migration := range planned {
		dur, err := d.driver.Exec(migration.Script)
		if err != nil {
			return err
		}

		if err := d.driver.Insert(MigrationRecord{
			Version:       migration.Version,
			Description:   migration.Description,
			Checksum:      migration.Checksum(),
			AppliedAt:     time.Now(),
			ExecutionTime: dur,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Info returns the status of all migrations.
func (d *Darwin) Info() ([]MigrationInfo, error) {
	records, err := d.driver.All()
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(byMigrationRecordVersion(records)))

	info := []MigrationInfo{}
	for _, migration := range d.migrations {
		info = append(info, MigrationInfo{
			Status:    getStatus(records, migration),
			Error:     nil,
			Migration: migration,
		})
	}

	return info, nil
}

// UpdateChecksums updates the checksum values in the migrations table for
// all existing migrations.
func (d *Darwin) UpdateChecksums() error {
	sort.Sort(byMigrationVersion(d.migrations))

	for _, migration := range d.migrations {
		d.driver.UpdateChecksum(migration.Checksum(), migration.Version)
	}

	return nil
}

// =============================================================================

func (d *Darwin) planMigration() ([]Migration, error) {
	records, err := d.driver.All()
	if err != nil {
		return []Migration{}, err
	}

	// Apply all migrations.
	if len(records) == 0 {
		return d.migrations, nil
	}

	// Which migrations needs to be applied.
	planned := []Migration{}

	// Make sure the order is correct. Do not trust the driver.
	sort.Sort(sort.Reverse(byMigrationRecordVersion(records)))
	last := records[0]

	// Apply all migrations that are greater than the last migration.
	for _, migration := range d.migrations {
		if migration.Version > last.Version {
			planned = append(planned, migration)
		}
	}

	// Make sure the order is correct.
	sort.Sort(byMigrationVersion(planned))

	return planned, nil
}

func (d *Darwin) isInvalidVersion() (float64, bool) {
	for _, migration := range d.migrations {
		version := migration.Version

		if version < 0 {
			return version, true
		}
	}

	return 0, false
}

func (d *Darwin) isDuplicated() (float64, bool) {
	unique := map[float64]Migration{}

	for _, migration := range d.migrations {
		_, exists := unique[migration.Version]

		if exists {
			return migration.Version, true
		}

		unique[migration.Version] = migration
	}

	return 0, false
}

func (d *Darwin) wasRemovedMigration(applied []MigrationRecord) (float64, bool) {
	versionMap := map[float64]Migration{}

	for _, migration := range d.migrations {
		versionMap[migration.Version] = migration
	}

	for _, migration := range applied {
		if _, ok := versionMap[migration.Version]; !ok {
			return migration.Version, true
		}
	}

	return 0, false
}

func (d *Darwin) isInvalidChecksumMigration(applied []MigrationRecord) (float64, bool) {
	versionMap := map[float64]MigrationRecord{}

	for _, migration := range applied {
		versionMap[migration.Version] = migration
	}

	for _, migration := range d.migrations {
		if m, ok := versionMap[migration.Version]; ok {
			if m.Checksum != migration.Checksum() {
				return migration.Version, true
			}
		}
	}

	return 0, false
}

// =============================================================================

func getStatus(inDatabase []MigrationRecord, migration Migration) Status {
	last := inDatabase[0]

	// Check if pending.
	if migration.Version > last.Version {
		return Pending
	}

	// Check if ignored.
	found := false

	for _, record := range inDatabase {
		if record.Version == migration.Version {
			found = true
		}
	}

	if !found {
		return Ignored
	}

	return Applied
}

// =============================================================================

type byMigrationVersion []Migration

func (b byMigrationVersion) Len() int           { return len(b) }
func (b byMigrationVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMigrationVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
