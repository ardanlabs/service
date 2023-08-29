package homedb

import "github.com/ardanlabs/service/business/core/home"

// dbHome represents an individual home.
type dbHome struct{}

// =============================================================================

func toDBHome(hme home.Home) dbHome {
	hmeDB := dbHome{}

	return hmeDB
}

func toCoreHome(dbHse dbHome) home.Home {
	hme := home.Home{}

	return hme
}

func toCoreHomeSlice(dbHomes []dbHome) []home.Home {
	hmes := make([]home.Home, len(dbHomes))
	for i, dbHse := range dbHomes {
		hmes[i] = toCoreHome(dbHse)
	}
	return hmes
}
