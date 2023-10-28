package darwin

import "fmt"

// DuplicateMigrationVersionError is used to report when the migration list has
// duplicated entries.
type DuplicateMigrationVersionError struct {
	Version float64
}

// Error implements the error interface.
func (d *DuplicateMigrationVersionError) Error() string {
	return fmt.Sprintf("Multiple migrations have the version number %f.", d.Version)
}

// IllegalMigrationVersionError is used to report when the migration has an
// illegal Version number.
type IllegalMigrationVersionError struct {
	Version float64
}

// Error implements the error interface.
func (i *IllegalMigrationVersionError) Error() string {
	return fmt.Sprintf("Illegal migration version number %f.", i.Version)
}

// RemovedMigrationError is used to report when a migration is removed from
// the list.
type RemovedMigrationError struct {
	Version float64
}

// Error implements the error interface.
func (r *RemovedMigrationError) Error() string {
	return fmt.Sprintf("Migration %f was removed", r.Version)
}

// InvalidChecksumError is used to report when a migration was modified.
type InvalidChecksumError struct {
	Version float64
}

// Error implements the error interface.
func (i *InvalidChecksumError) Error() string {
	return fmt.Sprintf("Invalid checksum for migration %f", i.Version)
}
