package homedb

import (
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/types/hometype"
	"github.com/google/uuid"
)

type home struct {
	ID          uuid.UUID `db:"home_id"`
	UserID      uuid.UUID `db:"user_id"`
	Type        string    `db:"type"`
	Address1    string    `db:"address_1"`
	Address2    string    `db:"address_2"`
	ZipCode     string    `db:"zip_code"`
	City        string    `db:"city"`
	Country     string    `db:"country"`
	State       string    `db:"state"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toDBHome(bus homebus.Home) home {
	db := home{
		ID:          bus.ID,
		UserID:      bus.UserID,
		Type:        bus.Type.String(),
		Address1:    bus.Address.Address1,
		Address2:    bus.Address.Address2,
		ZipCode:     bus.Address.ZipCode,
		City:        bus.Address.City,
		Country:     bus.Address.Country,
		State:       bus.Address.State,
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}

	return db
}

func toBusHome(db home) (homebus.Home, error) {
	typ, err := hometype.Parse(db.Type)
	if err != nil {
		return homebus.Home{}, fmt.Errorf("parse type: %w", err)
	}

	bus := homebus.Home{
		ID:     db.ID,
		UserID: db.UserID,
		Type:   typ,
		Address: homebus.Address{
			Address1: db.Address1,
			Address2: db.Address2,
			ZipCode:  db.ZipCode,
			City:     db.City,
			Country:  db.Country,
			State:    db.State,
		},
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toBusHomes(dbs []home) ([]homebus.Home, error) {
	bus := make([]homebus.Home, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toBusHome(db)
		if err != nil {
			return nil, fmt.Errorf("parse type: %w", err)
		}
	}

	return bus, nil
}
