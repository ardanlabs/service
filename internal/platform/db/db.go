// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0

package db

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
)

// ErrInvalidDBProvided is returned in the event that an uninitialized db is
// used to perform actions against.
var ErrInvalidDBProvided = errors.New("invalid DB provided")

// DB is a collection of support for different DB technologies. Currently
// only MongoDB has been implemented. We want to be able to access the raw
// database support for the given DB so an interface does not work. Each
// database is too different.
type DB struct {

	// MongoDB Support.
	database *mgo.Database
	session  *mgo.Session
}

// New returns a new DB value for use with MongoDB based on a registered
// master session.
func New(url string, timeout time.Duration) (*DB, error) {

	// Set the default timeout for the session.
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	ses, err := mgo.DialWithTimeout(url, timeout)
	if err != nil {
		return nil, errors.Wrapf(err, "mgo.DialWithTimeout: %s,%v", url, timeout)
	}

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	ses.SetMode(mgo.Monotonic, true)

	db := DB{
		database: ses.DB(""),
		session:  ses,
	}

	return &db, nil
}

// Close closes a DB value being used with MongoDB.
func (db *DB) Close() {
	db.session.Close()
}

// Copy returns a new DB value for use with MongoDB based on master session.
func (db *DB) Copy() *DB {
	ses := db.session.Copy()

	// As per the mgo documentation, https://godoc.org/gopkg.in/mgo.v2#Session.DB
	// if no database name is specified, then use the default one, or the one that
	// the connection was dialed with.
	newDB := DB{
		database: ses.DB(""),
		session:  ses,
	}

	return &newDB
}

// Execute is used to execute MongoDB commands.
func (db *DB) Execute(collName string, f func(*mgo.Collection) error) error {
	if db == nil || db.session == nil {
		return errors.Wrap(ErrInvalidDBProvided, "db == nil || db.session == nil")
	}

	return f(db.database.C(collName))
}

// ExecuteTimeout is used to execute MongoDB commands with a timeout.
func (db *DB) ExecuteTimeout(timeout time.Duration, collName string, f func(*mgo.Collection) error) error {
	if db == nil || db.session == nil {
		return errors.Wrap(ErrInvalidDBProvided, "db == nil || db.session == nil")
	}

	db.session.SetSocketTimeout(timeout)

	return f(db.database.C(collName))
}

// StatusCheck validates the DB status good.
func (db *DB) StatusCheck() error {
	return nil
}

// Query provides a string version of the value
func Query(value interface{}) string {
	json, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(json)
}
