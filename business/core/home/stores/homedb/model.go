package homedb

import (
	"time"

	"github.com/ardanlabs/service/business/core/home"
	"github.com/google/uuid"
)

// dbHome represents an individual home.
type dbHome struct {
	ID          uuid.UUID `db:"home_id"`
	Type        string    `db:"type"`
	UserID      uuid.UUID `db:"user_id"`
	Address1    string    `db:"address_1"`
	Address2    string    `db:"address_2"`
	ZipCode     string    `db:"zip_code"`
	City        string    `db:"city"`
	Country     string    `db:"country"`
	State       string    `db:"state"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

// =============================================================================

func toDBHome(hme home.Home) dbHome {
	hmeDB := dbHome{
		ID:          hme.ID,
		Type:        hme.Type,
		UserID:      hme.UserID,
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

func toCoreHome(dbHme dbHome) home.Home {
	hme := home.Home{
		ID:          dbHme.ID,
		Type:        dbHme.Type,
		UserID:      dbHme.UserID,
		DateCreated: dbHme.DateCreated.In(time.Local),
		DateUpdated: dbHme.DateUpdated.In(time.Local),
		Address: home.Address{
			Address1: dbHme.Address1,
			Address2: dbHme.Address2,
			ZipCode:  dbHme.ZipCode,
			City:     dbHme.City,
			Country:  dbHme.Country,
			State:    dbHme.State,
		},
	}

	return hme
}

func toCoreHomeSlice(dbHomes []dbHome) []home.Home {
	hmes := make([]home.Home, len(dbHomes))
	for i, dbHme := range dbHomes {
		hmes[i] = toCoreHome(dbHme)
	}
	return hmes
}
