package summarydb

import (
	"github.com/ardanlabs/service/business/cview/user/summary"
	"github.com/google/uuid"
)

type dbSummary struct {
	UserID     uuid.UUID `db:"user_id"`
	UserName   string    `db:"user_name"`
	TotalCount int       `db:"total_count"`
	TotalCost  float64   `db:"total_cost"`
}

func toCoreSummary(dbSmm dbSummary) summary.Summary {
	usr := summary.Summary{
		UserID:     dbSmm.UserID,
		UserName:   dbSmm.UserName,
		TotalCount: dbSmm.TotalCount,
		TotalCost:  dbSmm.TotalCost,
	}

	return usr
}

func toCoreSummarySlice(dbSmms []dbSummary) []summary.Summary {
	usrs := make([]summary.Summary, len(dbSmms))
	for i, dbSmm := range dbSmms {
		usrs[i] = toCoreSummary(dbSmm)
	}
	return usrs
}
