package usersummarygrp

import "github.com/ardanlabs/service/business/core/usersummary"

// AppUserSummary represents information about an individual user and their products.
type AppUserSummary struct {
	UserID     string  `json:"userID"`
	UserName   string  `json:"userName"`
	TotalCount int     `json:"totalCount"`
	TotalCost  float64 `json:"totalCost"`
}

func toAppUserSummary(smm usersummary.Summary) AppUserSummary {
	return AppUserSummary{
		UserID:     smm.UserID.String(),
		UserName:   smm.UserName,
		TotalCount: smm.TotalCount,
		TotalCost:  smm.TotalCost,
	}
}

func toAppUsersSummary(smms []usersummary.Summary) []AppUserSummary {
	items := make([]AppUserSummary, len(smms))
	for i, smm := range smms {
		items[i] = toAppUserSummary(smm)
	}

	return items
}
