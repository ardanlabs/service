package tests

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/homegrp"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/google/go-cmp/cmp"
)

func homeQuery200(sd seedData) []tableData {
	hmes := make([]home.Home, 0, len(sd.admins[0].homes)+len(sd.users[0].homes))
	hmes = append(hmes, sd.admins[0].homes...)
	hmes = append(hmes, sd.users[0].homes...)

	sort.Slice(hmes, func(i, j int) bool {
		return hmes[i].ID.String() <= hmes[j].ID.String()
	})

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/homes?page=1&rows=10&orderBy=home_id,ASC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &page.Document[homegrp.AppHome]{},
			expResp: &page.Document[homegrp.AppHome]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(hmes),
				Items:       toAppHomes(hmes),
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func homeQueryByID200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &homegrp.AppHome{},
			expResp:    toAppHomePtr(sd.users[0].homes[0]),
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
