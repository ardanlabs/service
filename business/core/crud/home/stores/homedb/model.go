package homedb

import (
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/google/uuid"
)

type dbHome struct {
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

func toDBHome(hme home.Home) dbHome {
	hmeDB := dbHome{
		ID:          hme.ID,
		UserID:      hme.UserID,
		Type:        hme.Type.Name(),
		Address1:    hme.Address.Address1,
		Address2:    hme.Address.Address2,
		ZipCode:     hme.Address.ZipCode,
		City:        hme.Address.City,
		Country:     hme.Address.Country,
		State:       hme.Address.State,
		DateCreated: hme.DateCreated.UTC(),
		DateUpdated: hme.DateUpdated.UTC(),
	}

	return hmeDB
}

func toCoreHome(dbHme dbHome) (home.Home, error) {
	typ, err := home.ParseType(dbHme.Type)
	if err != nil {
		return home.Home{}, fmt.Errorf("parse type: %w", err)
	}

	hme := home.Home{
		ID:     dbHme.ID,
		UserID: dbHme.UserID,
		Type:   typ,
		Address: home.Address{
			Address1: dbHme.Address1,
			Address2: dbHme.Address2,
			ZipCode:  dbHme.ZipCode,
			City:     dbHme.City,
			Country:  dbHme.Country,
			State:    dbHme.State,
		},
		DateCreated: dbHme.DateCreated.In(time.Local),
		DateUpdated: dbHme.DateUpdated.In(time.Local),
	}

	return hme, nil
}

func toCoreHomeSlice(dbHomes []dbHome) ([]home.Home, error) {
	hmes := make([]home.Home, len(dbHomes))

	for i, dbHme := range dbHomes {
		var err error
		hmes[i], err = toCoreHome(dbHme)
		if err != nil {
			return nil, fmt.Errorf("parse type: %w", err)
		}
	}

	return hmes, nil
}
