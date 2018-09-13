package product

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const productsCollection = "products"

var (
	// ErrNotFound abstracts the mgo not found error.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing products from the database.
func List(ctx context.Context, dbConn *db.DB) ([]Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.List")
	defer span.End()

	p := []Product{}

	f := func(collection *mgo.Collection) error {
		return collection.Find(nil).All(&p)
	}
	if err := dbConn.Execute(ctx, productsCollection, f); err != nil {
		return nil, errors.Wrap(err, "db.products.find()")
	}

	return p, nil
}

// Retrieve gets the specified product from the database.
func Retrieve(ctx context.Context, dbConn *db.DB, id string) (*Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.Retrieve")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return nil, ErrInvalidID
	}

	q := bson.M{"_id": bson.ObjectIdHex(id)}

	var p *Product
	f := func(collection *mgo.Collection) error {
		return collection.Find(q).One(&p)
	}
	if err := dbConn.Execute(ctx, productsCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, fmt.Sprintf("db.products.find(%s)", db.Query(q)))
	}

	return p, nil
}

// Create inserts a new product into the database.
func Create(ctx context.Context, dbConn *db.DB, cp *NewProduct, now time.Time) (*Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.Create")
	defer span.End()

	// Mongo truncates times to milliseconds when storing. We go ahead and do the
	// same here so the value we return is consistent with what we store.
	now = now.Truncate(time.Millisecond)

	p := Product{
		ID:           bson.NewObjectId(),
		Name:         cp.Name,
		Notes:        cp.Notes,
		UnitPrice:    cp.UnitPrice,
		Quantity:     cp.Quantity,
		Family:       cp.Family,
		DateCreated:  now,
		DateModified: now,
	}

	f := func(collection *mgo.Collection) error {
		return collection.Insert(&p)
	}
	if err := dbConn.Execute(ctx, productsCollection, f); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("db.products.insert(%s)", db.Query(&p)))
	}

	return &p, nil
}

// Update replaces a product document in the database.
func Update(ctx context.Context, dbConn *db.DB, id string, upd UpdateProduct, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.product.Update")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}

	fields := make(bson.M)

	if upd.Name != nil {
		fields["name"] = *upd.Name
	}
	if upd.Notes != nil {
		fields["notes"] = *upd.Notes
	}
	if upd.UnitPrice != nil {
		fields["unit_price"] = *upd.UnitPrice
	}
	if upd.Quantity != nil {
		fields["quantity"] = *upd.Quantity
	}
	if upd.Family != nil {
		fields["family"] = *upd.Family
	}

	// If there's nothing to update we can quit early.
	if len(fields) == 0 {
		return nil
	}

	fields["date_modified"] = now

	m := bson.M{"$set": fields}
	q := bson.M{"_id": bson.ObjectIdHex(id)}

	f := func(collection *mgo.Collection) error {
		return collection.Update(q, m)
	}
	if err := dbConn.Execute(ctx, productsCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return errors.Wrap(err, fmt.Sprintf("db.customers.update(%s, %s)", db.Query(q), db.Query(m)))
	}

	return nil
}

// Delete removes a product from the database.
func Delete(ctx context.Context, dbConn *db.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.product.Delete")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}

	q := bson.M{"_id": bson.ObjectIdHex(id)}

	f := func(collection *mgo.Collection) error {
		return collection.Remove(q)
	}
	if err := dbConn.Execute(ctx, productsCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			err = ErrNotFound
		}
		return errors.Wrap(err, fmt.Sprintf("db.products.remove(%v)", q))
	}

	return nil
}
