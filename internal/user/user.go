package user

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const usersCollection = "users"

// Create inserts a new user into the database.
func Create(ctx context.Context, dbConn *db.DB, cu *CreateUser) (*User, error) {
	now := time.Now()

	u := User{
		UserID:       bson.NewObjectId().Hex(),
		UserType:     cu.UserType,
		FirstName:    cu.FirstName,
		LastName:     cu.LastName,
		Email:        cu.Email,
		Company:      cu.Company,
		DateCreated:  &now,
		DateModified: &now,
	}

	f := func(collection *mgo.Collection) error {
		return collection.Insert(u)
	}
	if err := dbConn.Execute(ctx, usersCollection, f); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("db.users.insert(%s)", db.Query(u)))
	}

	return &u, nil
}
