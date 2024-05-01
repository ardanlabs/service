// Package transaction provides support for database transaction related
// functionality.
package transaction

// Beginner represents a value that can begin a transaction.
type Beginner interface {
	Begin() (CommitRollbacker, error)
}

// CommitRollbacker represents a value that can commit or rollback a transaction.
type CommitRollbacker interface {
	Commit() error
	Rollback() error
}
