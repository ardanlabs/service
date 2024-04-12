package tests

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/api/apptest"
	"github.com/ardanlabs/service/app/api/page"
	"github.com/ardanlabs/service/app/domain/homeapp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/google/go-cmp/cmp"
)

func homeQuery200(sd dbtest.SeedData) []apptest.AppTable {
	hmes := make([]homebus.Home, 0, len(sd.Admins[0].Homes)+len(sd.Users[0].Homes))
	hmes = append(hmes, sd.Admins[0].Homes...)
	hmes = append(hmes, sd.Users[0].Homes...)

	sort.Slice(hmes, func(i, j int) bool {
		return hmes[i].ID.String() <= hmes[j].ID.String()
	})

	table := []apptest.AppTable{
		{
			Name:       "basic",
			URL:        "/v1/homes?page=1&rows=10&orderBy=home_id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			Resp:       &page.Document[homeapp.Home]{},
			ExpResp: &page.Document[homeapp.Home]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(hmes),
				Items:       toAppHomes(hmes),
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func homeQueryByID200(sd dbtest.SeedData) []apptest.AppTable {
	table := []apptest.AppTable{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			Resp:       &homeapp.Home{},
			ExpResp:    toAppHomePtr(sd.Users[0].Homes[0]),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
